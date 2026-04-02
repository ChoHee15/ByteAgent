package greeting

import "strings"

func FormatGreeting(name string) string {
	cleaned := strings.TrimSpace(name)
	if cleaned == "" {
		return "Hello, stranger!"
	}

	return "Hello, " + strings.ToLower(cleaned) + "!"
}
