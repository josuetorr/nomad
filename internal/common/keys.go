package common

import (
	"fmt"
	"slices"
	"strings"
)

const (
	KeySep = "|"
)

func createkey(parts []string) string {
	var builder strings.Builder
	for i, p := range parts {
		if i == (len(parts) - 1) {
			builder.WriteString(p)
			continue
		}
		builder.WriteString(fmt.Sprintf("%s%s", p, KeySep))
	}

	return builder.String()
}

func TermKey(parts ...string) string {
	parts = slices.Insert(parts, 0, "term")
	return createkey(parts)
}

func DocKey(parts ...string) string {
	parts = slices.Insert(parts, 0, "doc")
	return createkey(parts)
}

func DFKey(parts ...string) string {
	parts = slices.Insert(parts, 0, "df")
	return createkey(parts)
}

func DocCountKey() string {
	return createkey([]string{"doc_count"})
}

func KeyParts(k string) []string {
	return strings.Split(k, KeySep)
}
