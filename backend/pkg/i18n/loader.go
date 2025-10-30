package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"sync"
)

//go:embed locales
var embeddedLocales embed.FS

func init() {
	InitMessageLoader()
}

// MessageLoader message loader
type MessageLoader struct {
	messages map[SupportedLanguage]map[int]string
	mu       sync.RWMutex
}

// NewMessageLoader creates a new message loader
func NewMessageLoader() *MessageLoader {
	return &MessageLoader{
		messages: make(map[SupportedLanguage]map[int]string),
	}
}

// LoadMessages 从嵌入文件系统加载所有支持的语言消息
func (ml *MessageLoader) LoadMessages() error {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	// 清空现有消息
	ml.messages = make(map[SupportedLanguage]map[int]string)

	// 从嵌入文件系统加载
	entries, err := embeddedLocales.ReadDir("locales")
	if err != nil {
		return fmt.Errorf("failed to read embedded locales directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 将目录名转换为支持的语言类型
		langCode := entry.Name()
		var lang SupportedLanguage
		switch langCode {
		case "zh-CN":
			lang = LanguageZhCN
		case "en-US":
			lang = LanguageEnUS
		default:
			// 跳过不支持的语言目录
			continue
		}

		langDir := "locales/" + langCode
		if err := ml.loadLanguageMessages(lang, langDir); err != nil {
			return fmt.Errorf("failed to load embedded messages for language %s: %w", lang, err)
		}
	}

	return nil
}

// loadLanguageMessages 从嵌入文件系统加载指定语言的所有消息文件
func (ml *MessageLoader) loadLanguageMessages(lang SupportedLanguage, langDir string) error {
	if ml.messages[lang] == nil {
		ml.messages[lang] = make(map[int]string)
	}

	// 遍历语言目录下的所有 JSON 文件
	return fs.WalkDir(embeddedLocales, langDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 只处理 JSON 文件
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		return ml.loadMessageFile(lang, path)
	})
}

// loadMessageFile 从嵌入文件系统加载单个消息文件
func (ml *MessageLoader) loadMessageFile(lang SupportedLanguage, filePath string) error {
	data, err := embeddedLocales.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read embedded file %s: %w", filePath, err)
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("failed to parse JSON file %s: %w", filePath, err)
	}

	// 将字符串键转换为整数键
	for codeStr, message := range messages {
		code, err := strconv.Atoi(codeStr)
		if err != nil {
			return fmt.Errorf("invalid error code '%s' in file %s: %w", codeStr, filePath, err)
		}
		ml.messages[lang][code] = message
	}

	return nil
}

// GetMessage 获取指定语言和错误码的消息
func (ml *MessageLoader) GetMessage(lang SupportedLanguage, code int) string {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	if messages, exists := ml.messages[lang]; exists {
		if message, found := messages[code]; found {
			return message
		}
	}

	// 如果指定语言没有找到，尝试使用英语作为回退
	if lang != LanguageEnUS {
		if messages, exists := ml.messages[LanguageEnUS]; exists {
			if message, found := messages[code]; found {
				return message
			}
		}
	}

	// 如果都没找到，返回默认消息
	return fmt.Sprintf("Unknown error code: %d", code)
}

// ReloadMessages 重新加载所有消息
func (ml *MessageLoader) ReloadMessages() error {
	return ml.LoadMessages()
}

// GetAllMessages 获取指定语言的所有消息
func (ml *MessageLoader) GetAllMessages(lang SupportedLanguage) map[int]string {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	if langMessages, exists := ml.messages[lang]; exists {
		// 返回副本以避免并发修改
		result := make(map[int]string)
		for code, message := range langMessages {
			result[code] = message
		}
		return result
	}

	return make(map[int]string)
}

// GetLoadedLanguages 获取已加载的语言列表
func (ml *MessageLoader) GetLoadedLanguages() []SupportedLanguage {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	var languages []SupportedLanguage
	for lang := range ml.messages {
		languages = append(languages, lang)
	}
	return languages
}

// AddSupportedLanguage 添加支持的语言
func (ml *MessageLoader) AddSupportedLanguage(lang SupportedLanguage) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if ml.messages[lang] == nil {
		ml.messages[lang] = make(map[int]string)
	}
}

// IsLanguageSupported 检查语言是否被支持
func (ml *MessageLoader) IsLanguageSupported(lang SupportedLanguage) bool {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	_, exists := ml.messages[lang]
	return exists
}

// 全局消息加载器实例
var globalMessageLoader *MessageLoader
var loaderOnce sync.Once

// InitMessageLoader 初始化全局消息加载器
func InitMessageLoader() error {
	var err error
	loaderOnce.Do(func() {
		globalMessageLoader = NewMessageLoader()
		err = globalMessageLoader.LoadMessages()
	})
	return err
}

// GetGlobalMessageLoader 获取全局消息加载器
func GetGlobalMessageLoader() *MessageLoader {
	if globalMessageLoader == nil {
		if err := InitMessageLoader(); err != nil {
			panic(fmt.Sprintf("failed to initialize global message loader: %v", err))
		}
	}
	return globalMessageLoader
}
