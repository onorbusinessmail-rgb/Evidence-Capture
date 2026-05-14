package multilang

// CurrentLanguage は現在の言語設定を保持します。
var CurrentLanguage = "jp"

// SetLanguage は言語コードを設定します。
// "English" が指定された場合は "en"、それ以外は "jp" とします。
func SetLanguage(lang string) {
	if lang == "English" {
		CurrentLanguage = "en"
	} else {
		CurrentLanguage = "jp"
	}
}

// T は現在の言語設定に基づいて翻訳テキストを返します。
func T(jp, en string) string {
	if CurrentLanguage == "en" {
		if en != "" {
			return en
		}
	}
	return jp
}
