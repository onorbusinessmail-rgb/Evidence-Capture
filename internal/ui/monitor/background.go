package uimonitor

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/ui/common"
	"time"
)

const (
	TimeoutCheckInterval = 1 * time.Second
	DiskCheckInterval    = 30 * time.Second
)

func (m *Monitor) StartBackgroundEnvironmentMonitor(stopCh <-chan struct{}, timeoutCh chan<- struct{}) {
	go func() {
		timeoutTicker := time.NewTicker(TimeoutCheckInterval)
		diskTicker := time.NewTicker(DiskCheckInterval)
		defer timeoutTicker.Stop()
		defer diskTicker.Stop()

		timeoutDuration := time.Duration(m.config.AutoSwitchTimeoutMinutes) * time.Minute
		if timeoutDuration <= 0 {
			timeoutDuration = 15 * time.Minute
		}

		for {
			select {
			case <-stopCh:
				return

			case <-timeoutTicker.C:
				if m.config.EnableAutoTimeout && time.Since(m.getLastActivityTime()) > timeoutDuration {
					select {
					case timeoutCh <- struct{}{}:
					default:
					}
					return
				}

			case <-diskTicker.C:
				m.CheckDiskFree()
			}
		}
	}()
}

func (m *Monitor) CheckDiskFree() {
	threshold := m.config.DiskFreeThresholdGB
	if threshold <= 0 {
		return
	}

	current := uicommon.GetCurrentDiskFree()
	isLow := current <= threshold

	if isLow && !m.isDiskWarningShown() {
		m.setDiskWarningShown(true)
		config.Log(
			"WARN", "空き容量が警告閾値以下です: 現在 %.2f GB / 閾値 %.2f GB", "Disk free space is low: current %.2f GB / threshold %.2f GB", current, threshold,
		)
	}

	if !isLow && m.isDiskWarningShown() {
		m.setDiskWarningShown(false)
	}
}

func (m *Monitor) getLastActivityTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastActivityTime
}

func (m *Monitor) setLastActivityTime(t time.Time) {
	m.mu.Lock()
	m.lastActivityTime = t
	m.mu.Unlock()
}

func (m *Monitor) isDiskWarningShown() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.diskWarningShown
}

func (m *Monitor) setDiskWarningShown(v bool) {
	m.mu.Lock()
	m.diskWarningShown = v
	m.mu.Unlock()
}
