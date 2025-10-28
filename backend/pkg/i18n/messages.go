package i18n

// getMessageTemplate gets message template from file system
func getMessageTemplate(lang SupportedLanguage, code int) string {
	loader := GetGlobalMessageLoader()
	return loader.GetMessage(lang, code)
}

// GetSupportedLanguages gets the list of supported languages
func GetSupportedLanguages() []SupportedLanguage {
	return []SupportedLanguage{LanguageZhCN, LanguageEnUS}
}
