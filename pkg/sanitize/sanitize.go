package sanitize

import "strings"

// Error скрывает чувствительные данные (например токен) в ошибках
func Error(err error, secret string) string {
	if err == nil {
		return ""
	}
	if secret == "" {
		return err.Error()
	}

	return strings.ReplaceAll(err.Error(), secret, "****")
}
