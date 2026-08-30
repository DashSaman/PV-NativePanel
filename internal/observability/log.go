package observability

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"time"
)

const redacted = "[REDACTED]"

type Event struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	RequestID string         `json:"request_id,omitempty"`
	Component string         `json:"component"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

var sensitiveKeys = map[string]struct{}{
	"authorization":     {},
	"dbpassword":        {},
	"dsn":               {},
	"password":          {},
	"privatekey":        {},
	"runtimesecret":     {},
	"secret":            {},
	"subscriptiontoken": {},
	"token":             {},
}

var redactionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(authorization\s*:\s*bearer\s+)[^\s,;]+`),
	regexp.MustCompile(`(?i)((?:subscription[_-]?token|db[_-]?password|runtime[_-]?secret|password|private[_-]?key)\s*[=:]\s*)[^\s,;]+`),
	regexp.MustCompile(`(?i)(postgres(?:ql)?://[^:\s/@]+:)[^@\s]+(@)`),
	regexp.MustCompile(`(?i)(naive\+https://[^:\s/@]+:)[^@\s]+(@)`),
	regexp.MustCompile(`(?i)(/(?:sub|s)/)[A-Za-z0-9_-]{32,}`),
	regexp.MustCompile(`(?i)(/api/v1/subscriptions/)[A-Za-z0-9_-]{32,}`),
}

func MarshalLog(event Event) ([]byte, error) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	} else {
		event.Timestamp = event.Timestamp.UTC()
	}
	if event.Level == "" {
		event.Level = "info"
	}
	event.Message = RedactText(event.Message)
	event.Fields = sanitizeFields(event.Fields)
	return json.Marshal(event)
}

func RedactText(value string) string {
	result := value
	for _, pattern := range redactionPatterns {
		if pattern.NumSubexp() == 2 {
			result = pattern.ReplaceAllString(result, `${1}`+redacted+`${2}`)
		} else {
			result = pattern.ReplaceAllString(result, `${1}`+redacted)
		}
	}
	return result
}

func sanitizeFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return nil
	}
	result := make(map[string]any, len(fields))
	for key, value := range fields {
		if isSensitiveKey(key) {
			result[key] = redacted
			continue
		}
		result[key] = sanitizeValue(value)
	}
	return result
}

func sanitizeValue(value any) any {
	switch typed := value.(type) {
	case string:
		return RedactText(typed)
	case []string:
		result := make([]string, len(typed))
		for i := range typed {
			result[i] = RedactText(typed[i])
		}
		return result
	case map[string]any:
		return sanitizeFields(typed)
	case []any:
		result := make([]any, len(typed))
		for i := range typed {
			result[i] = sanitizeValue(typed[i])
		}
		return result
	default:
		return value
	}
}

func isSensitiveKey(key string) bool {
	normalized := strings.NewReplacer("_", "", "-", "", ".", "", " ", "").Replace(strings.ToLower(key))
	if _, ok := sensitiveKeys[normalized]; ok {
		return true
	}
	for _, suffix := range []string{"password", "privatekey", "secret", "token"} {
		if strings.HasSuffix(normalized, suffix) {
			return true
		}
	}
	return false
}

func SensitiveFieldNames() []string {
	result := make([]string, 0, len(sensitiveKeys))
	for key := range sensitiveKeys {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
