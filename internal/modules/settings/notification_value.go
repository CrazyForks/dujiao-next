package settings

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/dujiao-next/internal/shared/jsonmap"
)

func toStringAnyMap(value interface{}) map[string]interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed
	case jsonmap.JSON:
		result := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			result[key] = item
		}
		return result
	default:
		return nil
	}
}

func readString(source map[string]interface{}, key, fallback string) string {
	value, ok := source[key]
	if !ok {
		return fallback
	}
	text, ok := value.(string)
	if !ok {
		return fallback
	}
	return strings.TrimSpace(text)
}

func readBool(source map[string]interface{}, key string, fallback bool) bool {
	value, ok := source[key]
	if !ok {
		return fallback
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return fallback
}

func readInt(source map[string]interface{}, key string, fallback int) int {
	value, ok := source[key]
	if !ok {
		return fallback
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int8:
		return int(typed)
	case int16:
		return int(typed)
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case uint:
		return int(typed)
	case uint8:
		return int(typed)
	case uint16:
		return int(typed)
	case uint32:
		return int(typed)
	case uint64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		if integer, err := typed.Int64(); err == nil {
			return int(integer)
		}
		if decimal, err := typed.Float64(); err == nil {
			return int(decimal)
		}
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
			return parsed
		}
	}
	return fallback
}

func cloneUintSlice(items []uint) []uint {
	if len(items) == 0 {
		return []uint{}
	}
	return append([]uint(nil), items...)
}

func readStringList(source map[string]interface{}, key string, fallback []string) []string {
	value, ok := source[key]
	if !ok {
		return cloneStringSlice(fallback)
	}
	switch typed := value.(type) {
	case []string:
		return cloneStringSlice(typed)
	case []interface{}:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return cloneStringSlice(fallback)
	}
}

func readUintList(source map[string]interface{}, key string, fallback []uint) []uint {
	value, ok := source[key]
	if !ok {
		return cloneUintSlice(fallback)
	}
	switch typed := value.(type) {
	case []uint:
		return cloneUintSlice(typed)
	case []interface{}:
		result := make([]uint, 0, len(typed))
		for _, item := range typed {
			switch number := item.(type) {
			case int:
				if number > 0 {
					result = append(result, uint(number))
				}
			case int64:
				if number > 0 {
					result = append(result, uint(number))
				}
			case uint:
				if number > 0 {
					result = append(result, number)
				}
			case float64:
				if number > 0 {
					result = append(result, uint(number))
				}
			}
		}
		return result
	default:
		return cloneUintSlice(fallback)
	}
}
