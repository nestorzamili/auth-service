package middleware

import (
	"strings"
)

func splitAndTrim(s, sep string) []string {
	parts := splitString(s, sep)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func splitString(s, sep string) []string {
	return strings.Split(s, sep)
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

func contains(slice []string, item string) bool {
	return indexOf(slice, item) != -1
}

func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}
