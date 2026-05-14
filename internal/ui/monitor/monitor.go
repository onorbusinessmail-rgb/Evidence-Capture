package uimonitor

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/excel"
	"Evidence-Capture/internal/types"
	"runtime"
	"sync"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/lxn/walk"
	"sync/atomic"
)

type Monitor struct {
	config           types.AppConfig
	captureInterval  time.Duration
	lastClickTimes   map[int]time.Time
	cornerStartTime  time.Time
	isInCorner       bool
	lastActivityTime time.Time
	recorder         *excel.Recorder

	mu               sync.RWMutex
	diskWarningShown bool
}

// 初期設定を行い構造体を生成
func NewMonitor(config types.AppConfig) *Monitor {
	return &Monitor{
		config:           config,
		captureInterval:  time.Duration(config.WaitTimeSeconds * float64(time.Second)),
		lastClickTimes:   make(map[int]time.Time),
		lastActivityTime: time.Now(),
		recorder:         excel.NewRecorder(config),
	}
}

const (
	CaptureChargeTime        = 500 * time.Millisecond
	PollingInterval          = 100 * time.Millisecond
	DblClickThreshold        = 700 * time.Millisecond
	KeyPressedMask           = 0x8000
	MONITOR_DEFAULTTONEAREST = 0x00000002
	MaxCellSearchRow         = 100000
)

// StartMonitoring は監視の自己修復コントローラーです
func StartMonitoring(mw *walk.MainWindow, isMonitoring *atomic.Bool, onLimitReached func()) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		if oleErr, ok := err.(*ole.OleError); ok && oleErr.Code() != 1 { // 1 = S_FALSE
			config.Log("ERROR", "StartMonitoring内でのCOM初期化に失敗しました: %v", "COM initialization failed in StartMonitoring: %v", err)
		}
	}
	defer ole.CoUninitialize()

	m := NewMonitor(config.CurrentConfig)

	for isMonitoring.Load() {
		m.RunPollingLoop(mw, isMonitoring, onLimitReached)

		if isMonitoring.Load() {
			config.Log("INFO", "入力が一定時間なかったため、監視機構をリフレッシュ（自己修復）しました。", "Monitor auto-refresh")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
