package service

import "strings"

// normalizeSettingText 统一清洗设置中的文本值。
func normalizeSettingText(raw interface{}) string {
	text, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

// normalizeSettingTextWithRuneLimit 清洗文本并限制最大字符数。
func normalizeSettingTextWithRuneLimit(raw interface{}, maxRuneCount int) string {
	text := normalizeSettingText(raw)
	if text == "" || maxRuneCount <= 0 {
		return text
	}

	runes := []rune(text)
	if len(runes) <= maxRuneCount {
		return text
	}
	return string(runes[:maxRuneCount])
}

// parseSettingBool 解析设置中的布尔值。
func parseSettingBool(raw interface{}) bool {
	switch value := raw.(type) {
	case bool:
		return value
	case int:
		return value != 0
	case int64:
		return value != 0
	case float64:
		return value != 0
	case string:
		normalized := strings.ToLower(strings.TrimSpace(value))
		return normalized == "1" || normalized == "true" || normalized == "yes" || normalized == "on"
	default:
		return false
	}
}

// cloneStringSlice 复制字符串切片，避免共享底层数组。
func cloneStringSlice(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	result := make([]string, len(items))
	copy(result, items)
	return result
}

// cloneUintSlice 复制无符号整数切片，避免共享底层数组。
func cloneUintSlice(items []uint) []uint {
	if len(items) == 0 {
		return []uint{}
	}
	result := make([]uint, len(items))
	copy(result, items)
	return result
}

// readStringList 从 map 中读取字符串列表，失败时回退默认值副本。
func readStringList(source map[string]interface{}, key string, fallback []string) []string {
	value, ok := source[key]
	if !ok {
		return cloneStringSlice(fallback)
	}
	switch raw := value.(type) {
	case []string:
		return cloneStringSlice(raw)
	case []interface{}:
		result := make([]string, 0, len(raw))
		for _, item := range raw {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return cloneStringSlice(fallback)
	}
}

// readUintList 从 map 中读取无符号整数列表，失败时回退默认值副本。
func readUintList(source map[string]interface{}, key string, fallback []uint) []uint {
	value, ok := source[key]
	if !ok {
		return cloneUintSlice(fallback)
	}
	switch raw := value.(type) {
	case []uint:
		return cloneUintSlice(raw)
	case []interface{}:
		result := make([]uint, 0, len(raw))
		for _, item := range raw {
			switch typed := item.(type) {
			case int:
				if typed > 0 {
					result = append(result, uint(typed))
				}
			case int64:
				if typed > 0 {
					result = append(result, uint(typed))
				}
			case uint:
				if typed > 0 {
					result = append(result, typed)
				}
			case float64:
				if typed > 0 {
					result = append(result, uint(typed))
				}
			}
		}
		return result
	default:
		return cloneUintSlice(fallback)
	}
}
