package common

import (
	"fmt"
	"slices"
	"strings"
)

const (
	keySep = ":"
)

func createkey(parts []string) string {
	var builder strings.Builder
	for i, p := range parts {
		if i == (len(parts) - 1) {
			builder.WriteString(p)
			continue
		}
		builder.WriteString(fmt.Sprintf("%s%s", p, keySep))
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

func KeyParts(k string) []string {
	return strings.Split(k, keySep)
}
