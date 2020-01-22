package lib

import "strings"

func containsInsensitive(larger, substring string) bool {
	return strings.Contains(strings.ToLower(larger), strings.ToLower(substring))
}
