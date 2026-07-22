package notification

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/queue"
)

var notificationTemplateVariablePattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

func buildNotificationTemplateVariables(payload queue.NotificationDispatchPayload) map[string]interface{} {
	data := cloneNotificationVariables(payload.Data)
	if data == nil {
		data = map[string]interface{}{}
	}
	data["event_type"] = strings.ToLower(strings.TrimSpace(payload.EventType))
	data["biz_type"] = strings.TrimSpace(payload.BizType)
	data["biz_id"] = fmt.Sprintf("%d", payload.BizID)
	data["occurred_at"] = time.Now().Format("2006-01-02 15:04:05")
	return data
}

func renderNotificationTemplate(tmpl string, variables map[string]interface{}) string {
	tmpl = strings.TrimSpace(tmpl)
	if tmpl == "" {
		return ""
	}
	return notificationTemplateVariablePattern.ReplaceAllStringFunc(tmpl, func(matched string) string {
		submatch := notificationTemplateVariablePattern.FindStringSubmatch(matched)
		if len(submatch) != 2 {
			return matched
		}
		value, ok := variables[strings.TrimSpace(submatch[1])]
		if !ok {
			return ""
		}
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	})
}

func resolveNotificationLocale(locale, fallback string) string {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		locale = strings.TrimSpace(fallback)
	}
	return normalizeNotificationLocale(locale)
}

func composeTelegramMessage(title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" {
		return body
	}
	if body == "" {
		return title
	}
	return title + "\n\n" + body
}

func notificationJSONToMap(data models.JSON) map[string]interface{} {
	if data == nil {
		return map[string]interface{}{}
	}
	result := make(map[string]interface{}, len(data))
	for key, value := range data {
		result[key] = value
	}
	return result
}

func cloneNotificationVariables(data map[string]interface{}) map[string]interface{} {
	if len(data) == 0 {
		return map[string]interface{}{}
	}
	result := make(map[string]interface{}, len(data))
	for key, value := range data {
		result[key] = value
	}
	return result
}
