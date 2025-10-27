package i18n

// getMessageTemplate 从文件系统获取消息模板
func getMessageTemplate(lang SupportedLanguage, code int) string {
	loader := GetGlobalMessageLoader()
	return loader.GetMessage(lang, code)
}

// GetSupportedLanguages 获取支持的语言列表
func GetSupportedLanguages() []SupportedLanguage {
	return []SupportedLanguage{LanguageZhCN, LanguageEnUS}
}
