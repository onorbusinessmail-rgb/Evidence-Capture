package ui

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/excel"
	"Evidence-Capture/internal/license"
	"Evidence-Capture/internal/multilang"
	"Evidence-Capture/internal/process"
	"Evidence-Capture/internal/support"
	"Evidence-Capture/internal/types"
	uicommon "Evidence-Capture/internal/ui/common"
	uiconfig "Evidence-Capture/internal/ui/config"
	uimonitor "Evidence-Capture/internal/ui/monitor"
	"Evidence-Capture/internal/winapi"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"sync/atomic"

	"github.com/go-ole/go-ole"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var (
	mw                 *walk.MainWindow
	icon               *walk.Icon
	startBtn           *walk.PushButton
	stopBtn            *walk.PushButton
	statusLabel        *walk.Label
	licenseStatusLabel *walk.Label
	isMonitoring       atomic.Bool
	nextInsertRow      int = 0

	// グループコンポーネント
	groupA             *walk.GroupBox
	groupB             *walk.GroupBox
	groupC             *walk.GroupBox
	pathTitleLabel     *walk.Label

	// ボタン類
	newExcelBtn        *walk.PushButton
	selectExcelBtn     *walk.PushButton
	localImgBtn        *walk.PushButton
	tocMgrBtn          *walk.PushButton
	indexMgrBtn        *walk.PushButton
	indexBtnMenuBtn    *walk.PushButton
	systemSettingsBtn  *walk.PushButton
	layoutSettingsBtn  *walk.PushButton
	languageBtn        *walk.PushButton
	licenseBtn         *walk.PushButton
	logOutputBtn       *walk.PushButton
	restartBtn         *walk.PushButton
	exitBtn            *walk.PushButton

	excelPathLabel     *walk.Label
	IsForceLimitMode   bool // mainから操作できるようにグローバルで定義
)

const (
	S_FALSE          = 1
	msoPicture       = 13
	msoLinkedPicture = 11
)

func StartApp() {
	config.Log("INFO", "ツール起動開始...", "Application started")
	runtime.LockOSThread()

	if err := config.CheckWritePermission(); err != nil {
		msg := fmt.Sprintf(
			"【起動エラー: 書き込み権限がありません】\n\n"+
				"現在のフォルダでは設定ファイルを保存できません。\n\n"+
				"■ 解決方法:\n"+
				"1. このツール（.exe）を、デスクトップやドキュメントフォルダなど\n"+
				"   書き込み権限のある場所に移動させてから再度実行してください。\n"+
				"2. または、ツールを右クリックして「管理者として実行」を試してください。\n\n"+
				"(詳細: %v)",
			err,
		)
		winapi.ShowDialog("権限エラー", msg, winapi.MB_ICONERROR)
		os.Exit(1)
	}

	if err := config.EnsureConfigFiles(); err != nil {
		msg := fmt.Sprintf("設定ファイルの作成に失敗しました。\nデスクトップ等、書き込み権限のある場所で実行してください。\n\n(詳細: %v)", err)
		winapi.ShowDialog("初期化エラー", msg, winapi.MB_ICONERROR)
		os.Exit(1)
	}

	if reflect.DeepEqual(config.CurrentConfig, types.AppConfig{}) {
		cfg, err := config.LoadAppConfig()
		if err != nil {
			winapi.ShowDialog("設定読み込みエラー", err.Error(), winapi.MB_ICONERROR)
			os.Exit(1)
		}
		config.CurrentConfig = cfg
	}

	oldMode := config.CurrentConfig.LicenseMode
	if err := license.InitializeLicense(&config.CurrentConfig); err == nil {
		if oldMode != "制限" && config.CurrentConfig.LicenseMode == "制限" {
			winapi.ShowDialog("ライセンス通知", "利用期限および回数が終了したため、制限モードに移行しました。", winapi.MB_ICONWARNING)
		}
	}

	// 1. 起動時の言語未設定判定
	if config.CurrentConfig.Language == "" {
		RunLanguageDialog(nil)
	}
	// 設定確定後に言語状態を同期
	multilang.SetLanguage(config.CurrentConfig.Language)

	handleExcelCleanupByConfig(config.CurrentConfig)

	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// TODO: 有償版向けライセンスキーチェック（現在は無償配布仕様のためコメントアウト）
	// isRestricted := IsForceLimitMode || !license.Check() || config.CurrentConfig.IsRestricted
	isRestricted := IsForceLimitMode || config.CurrentConfig.IsRestricted

	if IsForceLimitMode {
		config.CurrentConfig.IsRestricted = true
	}

	RunMainWindow(isRestricted)
}

func handleToolDuplicateInstance() uintptr {
	const mutexName = "MyEvidenceTool_Unique_Mutex_Name"
	handle, exists, err := process.CreateSingletonMutex(mutexName)
	if err != nil {
		config.Log("ERROR", "Mutexの作成に失敗しました: %v", "Failed to create mutex: %v", err)
		return 0
	}
	if exists {
		config.Log("ERROR", "既存のツールプロセスを検出しました。強制終了して入れ替えます。", "Existing process detected")
		os.Exit(0)
	}
	return handle
}

func RunMainWindow(isRestricted bool) {
	windowTitle := "Evidence-Capture"
	statusText := config.T("■ ステータス：待機中", "■ Status: Waiting")

	err := MainWindow{
		AssignTo: &mw,
		Title:    windowTitle,
		Icon:     2,
		Size:     Size{Width: 440, Height: 440},
		MinSize:  Size{Width: 320, Height: 200},
		Layout:   VBox{MarginsZero: false},
		Children: []Widget{
			VSpacer{Size: 10},
			Label{
				AssignTo:  &statusLabel,
				Text:      statusText,
				Font:      Font{Family: "Meiryo UI", PointSize: 11, Bold: true},
				Alignment: AlignHCenterVCenter,
				OnBoundsChanged: func() {
					// 画面表示時に現在のパスをセットし、ボタン状態を更新
					updateExcelPathDisplay(config.CurrentConfig.ExcelOutputPath)
					updateButtonStates()
				},
			},
			VSpacer{Size: 10},
			Label{
				Text:      config.T("⚠️ 制限モードで動作中", "⚠️ Running in Restricted Mode"),
				Visible:   IsForceLimitMode,
				Font:      Font{Family: "Meiryo UI", PointSize: 9, Bold: true},
				Alignment: AlignHCenterVCenter,
			},
			VSpacer{Size: 5},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &startBtn,
						Text:     config.T("証跡監視を開始", "Start Monitoring"),
						Font:     Font{Family: "Meiryo UI", PointSize: 12, Bold: true},
						MinSize:  Size{Width: 180, Height: 50},
						OnClicked: func() {
							if !isMonitoring.Load() {
								isMonitoring.Store(true)
								startBtn.SetEnabled(false)
								stopBtn.SetEnabled(true)
								statusLabel.SetText(config.T("■ ステータス：監視実行中", "■ Status: Monitoring"))

								if config.CurrentConfig.MinimizeExcel {
									excel.SetExcelWindowState(config.CurrentConfig.ExcelOutputPath, true)
								}
								go uimonitor.StartMonitoring(mw, &isMonitoring, func() {
									mw.Synchronize(func() {
										resetUI()
										UpdateLicenseDisplay()
									})
								})
							}
						},
					},
					HSpacer{Size: 10},
					PushButton{
						AssignTo: &stopBtn,
						Text:     config.T("停止", "Stop"),
						Enabled:  false,
						Font:     Font{Family: "Meiryo UI", PointSize: 10},
						MinSize:  Size{Width: 70, Height: 50},
						OnClicked: func() {
							isMonitoring.Store(false)
							resetUI()
							stopBtn.SetEnabled(false)
							config.Log("INFO", "監視を停止しました。", "Monitoring stopped")
						},
					},
					HSpacer{},
				},
			},
			VSpacer{Size: 10},
			Composite{
				Layout: VBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{
					Label{
						AssignTo: &pathTitleLabel,
						Text:     config.T("現在の保存先：", "Current Save Location:"),
						Font:     Font{Family: "Meiryo UI", PointSize: 9, Bold: true},
					},
					Composite{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							Label{
								AssignTo: &excelPathLabel,
								Text:     config.T("未設定", "Not Set"),
								Font:     Font{Family: "Meiryo UI", PointSize: 9},
							},
						},
					},
				},
			},
			VSpacer{Size: 10},

			// グループA：証跡ファイル操作
			GroupBox{
				AssignTo: &groupA,
				Title:    config.T("証跡ファイル操作", "Evidence File Operations"),
				Layout:   Grid{Columns: 3, Spacing: 10},
				Children: []Widget{
					menuButton(config.T("新規証跡ファイル作成", "Create New Evidence File"), &newExcelBtn, true, func() { openSubDialog(RunExcelNewManager) }),
					menuButton(config.T("既存のExcelを選択", "Select Existing Excel"), &selectExcelBtn, true, func() { selectExistingExcel() }),
					menuButton(config.T("画像のインポート管理", "Image Import Management"), &localImgBtn, true, func() { openSubDialog(RunImageImportDialog) }),
				},
			},

			// グループB：目次・進捗管理
			GroupBox{
				AssignTo: &groupB,
				Title:    config.T("目次・進捗管理", "TOC & Progress"),
				Layout:   Grid{Columns: 3, Spacing: 10},
				Children: []Widget{
					menuButton(config.T("目次管理メニュー", "Table of Contents Menu"), &tocMgrBtn, true, func() { openSubDialog(RunTOCManager) }),
					menuButton(config.T("インデックス管理", "Index Management"), &indexMgrBtn, config.CurrentConfig.MakeIndex, func() { openSubDialog(RunIndexManager) }),
					menuButton(config.T("ステータス更新パネル", "Status Panel"), &indexBtnMenuBtn, config.CurrentConfig.MakeIndex && excel.IsIndexSheetExists(), func() { openSubDialog(RunIndexButtonManager) }),
				},
			},

			// グループC：各種設定・管理（4列化）
			GroupBox{
				AssignTo: &groupC,
				Title:    config.T("各種設定・管理", "Settings & Management"),
				Layout:   Grid{Columns: 4, Spacing: 10},
				Children: []Widget{
					menuButton(config.T("システム設定", "System Settings"), &systemSettingsBtn, true, func() { openSubDialogWithAccept(uiconfig.RunConfigDialogEvidence) }),
					menuButton(config.T("レイアウト設定", "Layout Settings"), &layoutSettingsBtn, true, func() {
						openSubDialogWithAccept(func(f walk.Form, onA func()) { uiconfig.RunConfigDialogLayout(f, onA, 0) })
					}),
					menuButton(config.T("言語設定", "Language"), &languageBtn, true, func() { RunLanguageDialog(mw) }),
					menuButton(config.T("ライセンス認証", "License"), &licenseBtn, true, func() { openSubDialog(RunLicenseManager) }),
					menuButton(config.T("ログ出力", "Log Output"), &logOutputBtn, true, func() {
						now := time.Now().Format("20060102_150405")
						destFile := fmt.Sprintf("Diagnostic_Report_%s.zip", now)
						if err := support.CreateDiagnosticPackage(destFile); err != nil {
							uicommon.ShowErrorWithLogPrompt(mw, config.T("エラー", "Error"), config.T("診断情報の収集に失敗しました", "Failed to collect diagnostic info"))
							return
						}
						walk.MsgBox(mw, config.T("完了", "Completed"), config.T("診断用ZIPを出力しました", "Diagnostic ZIP exported"), walk.MsgBoxIconInformation)
						absPath, _ := filepath.Abs(destFile)
						winapi.ShowFileInExplorer(absPath)
					}),
					menuButton(config.T("再起動", "Restart"), &restartBtn, true, func() { restartTool() }),
					menuButton(config.T("終了", "Exit"), &exitBtn, true, func() { mw.Close() }),
				},
			},
			VSpacer{Size: 10},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					Label{AssignTo: &licenseStatusLabel, Font: Font{Family: "Meiryo UI", PointSize: 9}},
				},
			},
			Label{
				Text:      "Copyright (c) 2026 Evidence Solutions Lab. All Rights Reserved.",
				Font:      Font{PointSize: 8},
				TextColor: walk.RGB(128, 128, 128),
				Alignment: AlignHFarVNear,
			},
		},
	}.Create()

	if err != nil {
		log.Fatal(err)
	}

	// 起動時に最前面（アクティブ）にするための強力なアプローチ
	winapi.SetAlwaysOnTop(mw, true)
	winapi.SetAlwaysOnTop(mw, false)
	winapi.SetForegroundWindow.Call(uintptr(mw.Handle()))

	updateButtonStates()
	UpdateLicenseDisplay()

	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		ole.CoUninitialize()
	})

	mw.Run()
}

func UpdateLicenseDisplay() {
	if licenseStatusLabel == nil {
		return
	}
	conf := &config.CurrentConfig
	if conf.IsRestricted {
		if conf.NextAvailableTime != "" {
			t, err := time.Parse("2006-01-02 15:04:05", conf.NextAvailableTime)
			if err == nil {
				timeStr := t.Format("15:04")
				if conf.Language == "Japanese" || conf.Language == "" {
					timeStr = t.Format("15時04分")
				}
				licenseStatusLabel.SetText(fmt.Sprintf(config.T("制限モード動作中 (%s から再利用可能です)", "Restricted Mode (Available from %s)"), timeStr))
			} else {
				licenseStatusLabel.SetText(config.T("制限モード動作中 (待機中)", "Restricted Mode (Waiting)"))
			}
		} else {
			licenseStatusLabel.SetText(config.T("制限モード動作中", "Restricted Mode"))
		}
		licenseStatusLabel.SetTextColor(walk.RGB(255, 0, 0))
	} else {
		dateStr := conf.EndDate
		if dateStr == "" || dateStr == "--/--/--" {
			dateStr = config.T("未設定", "Not set")
		}
		licenseStatusLabel.SetText(fmt.Sprintf(config.T("残り利用可能回数: %d回 / 期限: %s", "Remaining: %d / Expires: %s"), conf.RemainingCount, dateStr))
		endDay, err := time.Parse("2006-01-02", conf.EndDate)
		daysLeft := 999
		if err == nil {
			daysLeft = int(time.Until(endDay).Hours() / 24)
		}
		if conf.RemainingCount <= 20 || daysLeft <= 30 {
			licenseStatusLabel.SetTextColor(walk.RGB(255, 0, 0))
		} else {
			licenseStatusLabel.SetTextColor(walk.RGB(128, 128, 128))
		}
	}
}

func resetUI() {
	isMonitoring.Store(false)
	excel.SetExcelWindowState(config.CurrentConfig.ExcelOutputPath, false)
	if startBtn != nil {
		startBtn.SetEnabled(true)
		startBtn.SetText(config.T("証跡監視を開始", "Start Monitoring"))
		config.Log("INFO", "証跡監視を開始しました", "resetUI startBtn is not nil")
	}
	if stopBtn != nil {
		stopBtn.SetEnabled(false)
	}
	if statusLabel != nil {
		statusLabel.SetText(config.T("■ ステータス：待機中", "■ Status: Waiting"))
	}
	updateButtonStates()
}

// updateButtonStates は現在の設定状況に応じてボタンの有効・無効および表示状態を切り替えます。
func updateButtonStates() {
	if mw == nil {
		return
	}
	conf := &config.CurrentConfig
	hasPath := conf.ExcelOutputPath != ""
	exists := excel.IsIndexSheetExists()

	// --- 1. 基本的な活性化制御 ---
	if startBtn != nil {
		startBtn.SetEnabled(hasPath)
	}
	if localImgBtn != nil {
		localImgBtn.SetEnabled(hasPath)
	}
	if tocMgrBtn != nil {
		tocMgrBtn.SetEnabled(hasPath)
	}

	// ログ設定によるボタン表示制御
	if logOutputBtn != nil {
		logOutputBtn.SetVisible(conf.EnableLogOutput)
	}

	// インデックス関連の制御
	if indexMgrBtn != nil {
		indexMgrBtn.SetEnabled(conf.MakeIndex && !exists)
	}
	if indexBtnMenuBtn != nil {
		isActive := hasActiveStatus()
		// ステータス定義が一つでも「可」であり、かつインデックスシートが存在する場合に有効化
		indexBtnMenuBtn.SetEnabled(isActive && exists)
		// 常に表示（無効状態でも存在は示す）
		indexBtnMenuBtn.SetVisible(true)
	}

	// --- 2. 制限モード（ライセンス）による上書き制御 ---
	if conf.IsRestricted {
		// 制限モードでは新規作成や画像挿入は原則不可
		if newExcelBtn != nil {
			newExcelBtn.SetEnabled(false)
		}
		if localImgBtn != nil {
			localImgBtn.SetEnabled(false)
		}

		isCooldown := false
		if conf.NextAvailableTime != "" {
			nextTime, err := time.Parse("2006-01-02 15:04:05", conf.NextAvailableTime)
			if err == nil && time.Now().Before(nextTime) {
				isCooldown = true
				if startBtn != nil {
					startBtn.SetEnabled(false)
				}
				if statusLabel != nil && !isMonitoring.Load() {
					timeStr := nextTime.Format("15:04")
					if conf.Language == "Japanese" || conf.Language == "" {
						timeStr = nextTime.Format("15時04分")
					}
					statusLabel.SetText(fmt.Sprintf(config.T("制限モード：%sまで待機中", "Restricted: Wait until %s"), timeStr))
				}
			}
		}
		if !isCooldown {
			// クールダウン中でなければ開始可能（ただしパス設定が必要）
			if startBtn != nil && !isMonitoring.Load() {
				startBtn.SetEnabled(hasPath)
			}
			if conf.SessionImageCount != 0 || conf.NextAvailableTime != "" {
				conf.SessionImageCount = 0
				conf.NextAvailableTime = ""
				license.SyncLicenseData(conf)
			}
		}
	}
}

func restartTool() {
	execPath, err := os.Executable()
	if err != nil {
		config.Log("ERROR", "実行ファイルのパスを取得できず、再起動を中止します %v", "Failed to get executable path %v", err)
		return
	}
	err = winapi.RestartSelf(execPath)
	if err != nil {
		config.Log("ERROR", "新しいプロセスを開始できず、再起動を中止します: %v", "Failed to start new process: %v", err)
		return
	}
	ole.CoUninitialize()
	os.Exit(0)
}

func handleExcelCleanupByConfig(cfg types.AppConfig) {
	config.Log("INFO", "Excelプロセス制御モード: %d", "Excel process control mode: %d", cfg.EnableMutex)
	switch cfg.EnableMutex {
	case 1:
		if cfg.ExcelOutputPath != "" {
			process.KillExcelProcesses(cfg.ExcelOutputPath)
		}
	case 2:
		process.KillExcelProcesses("")
	case 3:
	default:
	}
}

func openSubDialog(runDialogFunc func(walk.Form)) {
	runDialogFunc(mw)
}

func RunExcelNewManager(owner walk.Form) {
	newPath, err := excel.CreateAndRegisterNewExcel(owner)
	if err != nil {
		if err.Error() != "保存がキャンセルされました" {
			uicommon.ShowErrorWithLogPrompt(owner, config.T("エラー", "Error"), config.T("Excel作成に失敗しました:\n", "Failed to create Excel:\n")+err.Error())
		}
		return
	}
	// ダイアログを閉じた直後（MsgBoxの前）にUIを更新して反映
	refreshMainUI()
	walk.MsgBox(owner, config.T("成功", "Success"), config.T("新規Excelファイルを作成しました。\n", "Created new Excel file.\n")+newPath, walk.MsgBoxIconInformation)
}

func selectExistingExcel() {
	newPath, err := excel.SelectExistingExcel(mw)
	if err != nil {
		uicommon.ShowErrorWithLogPrompt(mw, config.T("エラー", "Error"), config.T("既存ファイルの選択に失敗しました", "Failed to select existing file"))
		return
	}
	if newPath != "" {
		refreshMainUI()
	}
}

func updateExcelPathDisplay(path string) {
	if excelPathLabel == nil {
		return
	}
	if path == "" {
		excelPathLabel.SetText(config.T("未設定", "Not Set"))
		excelPathLabel.SetToolTipText("")
		return
	}

	// フルパスをセット。レイアウト側で伸縮するように制御
	excelPathLabel.SetText(path)
	excelPathLabel.SetToolTipText(path) // ツールチップでフルパスを表示

	// パスが変わったのでボタンの状態も更新
	updateButtonStates()
}

// refreshMainUI は設定ファイル（INI）を再読み込みし、最新の値を使ってUIをリフレッシュします。
func refreshMainUI() {
	if cfg, err := config.LoadAppConfig(); err == nil {
		config.CurrentConfig = cfg
		updateButtonTexts() // テキストの多言語更新
		updateExcelPathDisplay(cfg.ExcelOutputPath)
		updateButtonStates()
		UpdateLicenseDisplay()
	}
}

// updateButtonTexts は現在の言語設定に基づいてUI上のすべてのテキストを更新します。
func updateButtonTexts() {
	if mw == nil {
		return
	}

	// メインステータス系
	if statusLabel != nil {
		if isMonitoring.Load() {
			statusLabel.SetText(config.T("■ ステータス：監視実行中", "■ Status: Monitoring"))
		} else {
			statusLabel.SetText(config.T("■ ステータス：待機中", "■ Status: Waiting"))
		}
	}
	if startBtn != nil {
		startBtn.SetText(config.T("証跡監視を開始", "Start Monitoring"))
	}
	if stopBtn != nil {
		stopBtn.SetText(config.T("停止", "Stop"))
	}
	if pathTitleLabel != nil {
		pathTitleLabel.SetText(config.T("現在の保存先：", "Current Save Location:"))
	}

	// グループA
	if groupA != nil {
		groupA.SetTitle(config.T("証跡ファイル操作", "Evidence File Operations"))
	}
	if newExcelBtn != nil {
		newExcelBtn.SetText(config.T("新規証跡ファイル作成", "Create New Evidence File"))
	}
	if selectExcelBtn != nil {
		selectExcelBtn.SetText(config.T("既存のExcelを選択", "Select Existing Excel"))
	}
	if localImgBtn != nil {
		localImgBtn.SetText(config.T("画像のインポート管理", "Image Import Management"))
	}

	// グループB
	if groupB != nil {
		groupB.SetTitle(config.T("目次・進捗管理", "TOC & Progress"))
	}
	if tocMgrBtn != nil {
		tocMgrBtn.SetText(config.T("目次管理メニュー", "Table of Contents Menu"))
	}
	if indexMgrBtn != nil {
		indexMgrBtn.SetText(config.T("インデックス管理", "Index Management"))
	}
	if indexBtnMenuBtn != nil {
		indexBtnMenuBtn.SetText(config.T("ステータス更新パネル", "Status Panel"))
	}

	// グループC
	if groupC != nil {
		groupC.SetTitle(config.T("各種設定・管理", "Settings & Management"))
	}
	if systemSettingsBtn != nil {
		systemSettingsBtn.SetText(config.T("システム設定", "System Settings"))
	}
	if layoutSettingsBtn != nil {
		layoutSettingsBtn.SetText(config.T("レイアウト設定", "Layout Settings"))
	}
	if languageBtn != nil {
		languageBtn.SetText(config.T("言語設定", "Language"))
	}
	if licenseBtn != nil {
		licenseBtn.SetText(config.T("ライセンス認証", "License"))
	}
	if logOutputBtn != nil {
		logOutputBtn.SetText(config.T("ログ出力", "Log Output"))
	}
	if restartBtn != nil {
		restartBtn.SetText(config.T("再起動", "Restart"))
	}
	if exitBtn != nil {
		exitBtn.SetText(config.T("終了", "Exit"))
	}
}

func openSubDialogWithAccept(runDialogFunc func(walk.Form, func())) {
	runDialogFunc(mw, func() {
		// 設定から戻った際に最新の設定を読み込み、UIをリフレッシュ
		refreshMainUI()
	})
}

func menuButton(text string, assignTo **walk.PushButton, enabled bool, onClick func()) PushButton {
	pb := PushButton{
		Text:      text,
		Font:      Font{Family: "Meiryo UI", PointSize: 9},
		MinSize:   Size{Width: 110, Height: 35},
		Enabled:   enabled,
		OnClicked: onClick,
	}
	if assignTo != nil {
		pb.AssignTo = assignTo
	}
	return pb
}

// hasActiveStatus は現在の設定に有効なステータス定義が一つでもあるか判定します。
func hasActiveStatus() bool {
	for _, val := range config.CurrentConfig.StatusDefinitions {
		parts := strings.Split(val, "|")
		// 4項目目が存在し、かつそれが「可」または「ON」である場合のみ有効とみなす
		if len(parts) >= 4 {
			st := strings.TrimSpace(parts[3])
			if st == "可" || st == "ON" {
				return true
			}
		}
	}
	return false
}
