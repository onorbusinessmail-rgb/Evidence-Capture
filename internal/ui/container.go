package ui

import (
	"Evidence-Capture/internal/types"

	"github.com/lxn/walk"
)

type UIContainer struct {
	MainWindow *walk.MainWindow
	Config     *types.AppConfig
}
