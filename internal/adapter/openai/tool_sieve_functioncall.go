package openai

import "strings"

func findQuotedFunctionCallKeyStart(s string) int {
	lower := strings.ToLower(s)
	const key = "\"functioncall\""
	for from := 0; from < len(lower); {
		rel := strings.Index(lower[from:], key)
		if rel < 0 {
			return -1
		}
		idx := from + rel
		if !hasJSONObjectContextPrefix(lower[:idx]) {
			from = idx + 1
			continue
		}
		j := idx + len(key)
		for j < len(lower) && (lower[j] == ' ' || lower[j] == '\t' || lower[j] == '\r' || lower[j] == '\n') {
			j++
		}
		if j < len(lower) && lower[j] == ':' {
			return idx
		}
		from = idx + 1
	}
	return -1
}

func hasJSONObjectContextPrefix(prefix string) bool {
	return strings.LastIndex(prefix, "{") >= 0
}
