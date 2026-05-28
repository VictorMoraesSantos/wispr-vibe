package processor

import (
	"regexp"
	"strings"
	"unicode"
)

var fillerPatterns = regexp.MustCompile(`(?i)\b(uh+|hm+|ah+|eh+|tipo assim|então né|né\b|you know|like\b|basically|actually|so yeah)\b`)
var fillerUm = regexp.MustCompile(`(?i)\bum\b`)

func RemoveFillers(text string) string {
	text = fillerPatterns.ReplaceAllString(text, "")
	text = fillerUm.ReplaceAllString(text, "")
	return text
}

var multiSpace = regexp.MustCompile(`\s{2,}`)

func CollapseSpaces(text string) string {
	return multiSpace.ReplaceAllString(text, " ")
}

func TrimText(text string) string {
	return strings.TrimSpace(text)
}

func CapitalizeFirst(text string) string {
	if text == "" {
		return text
	}
	runes := []rune(text)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func EnsurePeriod(text string) string {
	if text == "" {
		return text
	}
	last := text[len(text)-1]
	if last != '.' && last != '!' && last != '?' && last != ':' && last != ';' {
		return text + "."
	}
	return text
}
