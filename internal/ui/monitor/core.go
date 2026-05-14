package uimonitor

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/license"
	"Evidence-Capture/internal/process"
	"Evidence-Capture/internal/winapi"
	"fmt"
	"time"

	"github.com/lxn/walk"
	"sync/atomic"
)

// RunPollingLoop は1回分の監視ループを実行します。
func (m *Monitor) RunPollingLoop(mw *walk.MainWindow, isMonitoring *atomic.Bool, onLimitReached func()) {
	// リフレッシュトラッキング用変数
	lastMousePt := m.GetCurrentCursorPos()
	lastMouseMoveTime := time.Now()
	refreshThreshold := 3 * time.Minute // 3分間入力・変化がなければリフレッシュ

	// バックグラウンドでタイムアウト監視を開始（別ゴルーチン）
	stopTimeoutMonitor := m.StartTimeoutMonitor(mw, isMonitoring)
	defer stopTimeoutMonitor()

	for isMonitoring.Load() {
		pt := m.GetCurrentCursorPos()

		// 1. 監視機能の自己修復（リフレッシュ）判定
		if pt.X != lastMousePt.X || pt.Y != lastMousePt.Y {
			lastMousePt = pt
			lastMouseMoveTime = time.Now()
		} else if time.Since(lastMouseMoveTime) > refreshThreshold {
			return
		}

		// 2. 終了トリガー判定
		if m.ShouldExit(pt) {
			isMonitoring.Store(false)
			// Callback needed for ResetUI
			if onLimitReached != nil {
				onLimitReached()
			}
			return
		}

		// 3. 撮影トリガー判定
		if m.ShouldCapture(pt) {
			m.PerformCapture(mw, isMonitoring, onLimitReached)
			time.Sleep(m.captureInterval)
			lastMouseMoveTime = time.Now()
		}

		time.Sleep(PollingInterval)
	}
}

// StartTimeoutMonitor はタイムアウト判定を別ゴルーチンで実行します。
func (m *Monitor) StartTimeoutMonitor(mw *walk.MainWindow, isMonitoring *atomic.Bool) func() {
	if !m.config.EnableAutoTimeout {
		return func() {}
	}

	timeoutDuration := time.Duration(m.config.AutoSwitchTimeoutMinutes) * time.Minute
	if timeoutDuration <= 0 {
		timeoutDuration = 15 * time.Minute
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if time.Since(m.lastActivityTime) > timeoutDuration {
					isMonitoring.Store(false)
					mw.Synchronize(func() {
						msg := fmt.Sprintf("%d分間操作がなかったため、監視を自動終了しました。", m.config.AutoSwitchTimeoutMinutes)
						winapi.ShowDialog("タイムアウト", msg, winapi.MB_ICONINFORMATION)
						config.Log("INFO", "タイムアウトによる終了", "Timeout")
					})
					return
				}
			case <-stopCh:
				return
			}
		}
	}()

	return func() {
		close(stopCh)
		<-doneCh
	}
}

// PerformCapture は撮影実行を行います
func (m *Monitor) PerformCapture(mw *walk.MainWindow, isMonitoring *atomic.Bool, onLimitReached func()) {
	captureTime := time.Now()
	conf := &config.CurrentConfig

	// 1. 制限モード時の事前チェック
	if conf.IsRestricted {
		if conf.NextAvailableTime != "" {
			nextTime, err := time.Parse("2006-01-02 15:04:05", conf.NextAvailableTime)
			if err == nil && time.Now().Before(nextTime) {
				m.PlayFailureBeep()
				config.Log("WARN", "制限モードの待機時間中のため撮影をブロックしました", "Capture blocked due to restricted mode cooldown")
				return
			}
		}

		conf.SessionImageCount++
		license.SyncLicenseData(conf)

		if conf.SessionImageCount > 5 {
			cooldownEndTime := time.Now().Add(30 * time.Minute)
			conf.NextAvailableTime = cooldownEndTime.Format("2006-01-02 15:04:05")
			license.SyncLicenseData(conf)

			timeStr := cooldownEndTime.Format("15:04")
			if config.CurrentConfig.Language == "Japanese" || config.CurrentConfig.Language == "" {
				timeStr = cooldownEndTime.Format("15時04分")
			}
			msg := fmt.Sprintf(config.T("規定の5枚に達したため、撮影を制限しました。\n次回利用開始時刻は %s です\n\nOKを押すと監視を終了します", "Capture limit of 5 reached. Restricted mode activated.\nNext available time is %s\n\nClick OK to stop monitoring"), timeStr)

			m.PlayFailureBeep()
			config.Log("INFO", "制限枚数に達したため、%s までロックしました", "Locked due to limit. Cooldown until %s", timeStr)

			if mw != nil {
				mw.Synchronize(func() {
					winapi.ShowDialog(config.T("制限モード", "Restricted Mode"), msg, winapi.MB_ICONWARNING)
					isMonitoring.Store(false)

					// 制限枚数到達時にアプリを自己再起動して状態をリセットする
					config.Log("INFO", "制限枚数到達によりアプリを再起動します", "Restarting app due to limit reached")
					process.RestartSelf()
				})
			}
			return
		}
	}

	// 2. 実際のキャプチャ処理
	if err := m.CaptureActiveWindowRobust(); err != nil {
		config.Log("ERROR", "スクリーンショット取得失敗: %v", "Failed to capture screenshot: %v", err)
		m.PlayFailureBeep()
		return
	}

	m.PlaySuccessBeep()
	time.Sleep(300 * time.Millisecond)

	if err := m.recorder.ProcessCapture(captureTime); err != nil {
		config.Log("ERROR", "Excel貼り付け失敗: %v", "Failed to paste to Excel: %v", err)
		m.PlayFailureBeep()
	} else {
		m.setLastActivityTime(captureTime)
	}
}

func (m *Monitor) PlaySuccessBeep() {
	if m.config.EnableBeep {
		winapi.PlayBeep()
	}
}

func (m *Monitor) PlayFailureBeep() {
	if m.config.EnableBeep {
		winapi.PlayBeep()
		time.Sleep(150 * time.Millisecond)
		winapi.PlayBeep()
	}
}

func (m *Monitor) CanOpenClipboard() bool {
	ret, _, _ := winapi.OpenClipboard.Call(0)
	if ret == 0 {
		return false
	}
	winapi.CloseClipboard.Call()
	return true
}
