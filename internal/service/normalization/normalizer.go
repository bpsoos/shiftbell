package normalization

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type Config struct {
	NameLimit        int
	DescriptionLimit int
	SearchLimit      int
}

type Normalizer struct {
	nameLimit        int
	descriptionLimit int
	searchLimit      int
}

func New(config Config) *Normalizer {
	return &Normalizer{
		nameLimit:        config.NameLimit,
		descriptionLimit: config.DescriptionLimit,
		searchLimit:      config.SearchLimit,
	}
}

func (n *Normalizer) NormalizeName(value string) (string, bool) {
	if !utf8.ValidString(value) {
		return "", false
	}

	value = norm.NFC.String(strings.TrimSpace(value))
	if value == "" || utf8.RuneCountInString(value) > n.nameLimit {
		return "", false
	}

	for _, r := range value {
		if !unicode.IsPrint(r) {
			return "", false
		}
	}

	return value, true
}

func (n *Normalizer) NormalizeDescription(value string) (string, bool) {
	if !utf8.ValidString(value) {
		return "", false
	}

	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = norm.NFC.String(value)
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > n.descriptionLimit {
		return "", false
	}

	for _, r := range value {
		if !unicode.IsPrint(r) && r != '\t' && r != '\n' {
			return "", false
		}
	}

	return value, true
}

func (n *Normalizer) NormalizeSearch(value string) (string, bool) {
	if !utf8.ValidString(value) {
		return "", false
	}

	value = norm.NFC.String(strings.TrimSpace(value))
	if utf8.RuneCountInString(value) > n.searchLimit {
		return "", false
	}

	for _, r := range value {
		if !unicode.IsPrint(r) {
			return "", false
		}
	}

	return value, true
}
