package config

// 汎用的なエラー・ログメッセージの定数定義

const (
	// ファイル操作関連
	MsgErrOpenFileEn   = "Failed to open file"
	MsgErrOpenFileJa   = "ファイルを開けませんでした"
	MsgErrSaveFileEn   = "Failed to save file"
	MsgErrSaveFileJa   = "ファイルの保存に失敗しました"
	MsgErrCreateFileEn = "Failed to create file"
	MsgErrCreateFileJa = "ファイルの作成に失敗しました"

	// パース・フォーマット関連
	MsgErrParseFormatEn = "Format parsing failed"
	MsgErrParseFormatJa = "形式の解析に失敗しました"
	MsgErrDecodeEn      = "Decoding failed"
	MsgErrDecodeJa      = "デコードに失敗しました"
	MsgErrEncodeEn      = "Encoding failed"
	MsgErrEncodeJa      = "エンコードに失敗しました"

	// ダイアログ・UI関連
	MsgErrCreateDialogEn = "Failed to create dialog"
	MsgErrCreateDialogJa = "ダイアログの作成に失敗しました"

	// プロセス・システム関連
	MsgErrProcessEn = "Process error occurred"
	MsgErrProcessJa = "プロセスエラーが発生しました"
	MsgErrMutexEn   = "Failed to create mutex"
	MsgErrMutexJa   = "Mutexの作成に失敗しました"

	// 汎用エラー
	MsgErrGeneralEn = "An error occurred"
	MsgErrGeneralJa = "エラーが発生しました"
)
