package uimonitor

import (
	"Evidence-Capture/internal/winapi"
	"time"
	"unsafe"
)

// ShouldExit 終了トリガー判定
func (m *Monitor) ShouldExit(pt winapi.POINT) bool {
	if m.config.EnableExitShortcut && IsKeyPressed(winapi.VK_CONTROL) && IsKeyPressed(winapi.VK_SHIFT) && IsKeyPressed(winapi.VK_X) {
		return true
	}
	if m.config.EnableExitWinKey && (IsKeyPressed(winapi.VK_LWIN) || IsKeyPressed(winapi.VK_RWIN)) {
		return true
	}
	if m.config.EnableExitLeftDbl && isAtLeftCorner(pt, m.config.EdgeSensitivity) && m.IsDoubleClickDetected(winapi.VK_LBUTTON) {
		return true
	}
	return false
}

// ShouldCapture 撮影トリガー判定
func (m *Monitor) ShouldCapture(pt winapi.POINT) bool {
	if m.config.EnableTriggerRightDbl && m.IsDoubleClickDetected(winapi.VK_RBUTTON) {
		return true
	}

	isMouseOver := isMouseAtTopEdge(pt, m.config.EdgeSensitivity) &&
		((m.config.EnableTriggerTopLeft && isAtLeftCorner(pt, m.config.EdgeSensitivity)) ||
			(m.config.EnableTriggerTopRight && isAtRightCorner(pt, m.config.EdgeSensitivity)))

	if isMouseOver {
		if !m.isInCorner {
			m.cornerStartTime = time.Now()
			m.isInCorner = true
		} else if time.Since(m.cornerStartTime) >= CaptureChargeTime {
			m.isInCorner = false
			return true
		}
	} else {
		m.isInCorner = false
	}

	return false
}

// IsDoubleClickDetected ヘルパー：ダブルクリック判定
func (m *Monitor) IsDoubleClickDetected(vkey int) bool {
	if IsKeyPressed(vkey) {
		now := time.Now()
		lastTime := m.lastClickTimes[vkey]

		if now.Sub(lastTime) < DblClickThreshold && now.Sub(lastTime) > 100*time.Millisecond {
			m.lastClickTimes[vkey] = time.Time{}
			return true
		}
		m.lastClickTimes[vkey] = now
	}
	return false
}

// IsKeyPressed ヘルパー：キー押下判定
func IsKeyPressed(vkey int) bool {
	ret, _, _ := winapi.GetAsyncKeyState.Call(uintptr(vkey))
	return ret&KeyPressedMask != 0
}

// GetCurrentCursorPos ヘルパー：マウスカーソルの現在位置を取得
func (m *Monitor) GetCurrentCursorPos() winapi.POINT {
	var pt winapi.POINT
	winapi.GetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return pt
}

func getMonitorRectAtPoint(pt winapi.POINT) winapi.RECT {
	hMonitor, _, _ := winapi.MonitorFromPoint.Call(uintptr(pt.X), uintptr(pt.Y), 2)
	var mi winapi.MONITORINFO
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	winapi.GetMonitorInfo.Call(hMonitor, uintptr(unsafe.Pointer(&mi)))
	return mi.RcMonitor
}

func isMouseAtTopEdge(pt winapi.POINT, sensitivity float64) bool {
	rect := getMonitorRectAtPoint(pt)
	margin := int32(float64(rect.Right-rect.Left) * sensitivity)
	return pt.Y <= rect.Top+margin
}

func isAtLeftCorner(pt winapi.POINT, sensitivity float64) bool {
	rect := getMonitorRectAtPoint(pt)
	width := rect.Right - rect.Left
	if width <= 0 {
		return false
	}
	margin := int32(float64(width) * sensitivity)
	// モニター内の相対座標で判定
	return pt.X <= (rect.Left + margin)
}

func isAtRightCorner(pt winapi.POINT, sensitivity float64) bool {
	rect := getMonitorRectAtPoint(pt)
	width := rect.Right - rect.Left
	if width <= 0 {
		return false
	}
	margin := int32(float64(width) * sensitivity)
	// モニター内の相対座標で判定
	return pt.X >= (rect.Right - margin)
}
