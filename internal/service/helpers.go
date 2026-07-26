package service

import (
	"strings"
)

// pickFirstNonEmpty 返回第一个非空（trim 后）的字符串。
func pickFirstNonEmpty(values ...string) string {
	for _, val := range values {
		trimmed := strings.TrimSpace(val)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
