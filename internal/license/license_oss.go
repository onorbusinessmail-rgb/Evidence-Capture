//go:build !pro

package license

import "Evidence-Capture/internal/types"

// OSS版は常にライセンスチェックをスキップ（有効）とする
func Check() bool {
	return true
}

func InitializeLicense(conf *types.AppConfig) error {
	conf.IsRestricted = false
	conf.LicenseMode = "OSS (無償版)"
	conf.RemainingCount = 9999
	return nil
}

func SyncLicenseData(conf *types.AppConfig) {
	// OSS版は何もしない
}

func GetMachineID() string {
	return "OSS-USER-MACHINE"
}
