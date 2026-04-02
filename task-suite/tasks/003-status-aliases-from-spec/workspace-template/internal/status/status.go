package status

import "strings"

func NormalizeStatus(input string) string {
	normalized := strings.ToLower(strings.TrimSpace(input))

	switch normalized {
	case "active":
		return "active"
	case "inactive":
		return "inactive"
	default:
		return "unknown"
	}
}
