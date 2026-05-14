// --- internal/excel/utils.go (新規) ---
package excel

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseCellAddress はセル番地を分解する共通関数
func ParseCellAddress(cellStr string) (col string, row int, err error) {
	re := regexp.MustCompile(`^([A-Z]+)([0-9]+)$`)
	matches := re.FindStringSubmatch(strings.ToUpper(cellStr))
	if len(matches) == 3 {
		col = matches[1]
		fmt.Sscanf(matches[2], "%d", &row)
		return col, row, nil
	}
	return "B", 4, fmt.Errorf("invalid format")
}

// HexToExcelColor は色変換ロジックを分離
func HexToExcelColor(hex string) (int64, error) {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return 0, err
	}
	return int64(b)<<16 | int64(g)<<8 | int64(r), nil
}
