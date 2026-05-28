package processor

import (
	"regexp"
	"strings"
	"unicode"
)

// Common filler words in PT-BR and EN — use word boundaries carefully
var fillerPatterns = regexp.MustCompile(`(?i)\b(uh+|hm+|ah+|eh+|tipo assim|então né|né\b|you know|like\b|basically|actually|so yeah)\b`)
var fillerUm = regexp.MustCompile(`(?i)\bum\b`)

// RemoveFillers strips common filler/hesitation words.
func RemoveFillers(text string) string {
	text = fillerPatterns.ReplaceAllString(text, "")
	text = fillerUm.ReplaceAllString(text, "")
	return text
}

var multiSpace = regexp.MustCompile(`\s{2,}`)

// CollapseSpaces reduces multiple whitespace to single space.
func CollapseSpaces(text string) string {
	return multiSpace.ReplaceAllString(text, " ")
}

// TrimText trims leading/trailing whitespace.
func TrimText(text string) string {
	return strings.TrimSpace(text)
}

// CapitalizeFirst capitalizes the first character.
func CapitalizeFirst(text string) string {
	if text == "" {
		return text
	}
	runes := []rune(text)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// EnsurePeriod adds a period at the end if no punctuation present.
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
