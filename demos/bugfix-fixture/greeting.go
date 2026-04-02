package greeting

import "strings"

// FormatGreeting returns a simple greeting for the provided name.
func FormatGreeting(name string) string {
	cleaned := strings.TrimSpace(name)
	if cleaned == "" {
		return "Hello, stranger!"
	}

	return "Hello, " + strings.ToLower(cleaned) + "!"
}
