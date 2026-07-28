package normalization

import (
	"strings"
	"unicode"
	"unicode/utf8"

	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
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

func (n *Normalizer) NormalizeName(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", validationerrors.ErrInvalidUTF8
	}

	value = norm.NFC.String(strings.TrimSpace(value))
	if value == "" {
		return "", validationerrors.ErrRequired
	}
	if utf8.RuneCountInString(value) > n.nameLimit {
		return "", validationerrors.ErrTooLong
	}

	for _, r := range value {
		if !unicode.IsPrint(r) {
			return "", validationerrors.ErrDisallowedCharacter
		}
	}

	return value, nil
}

func (n *Normalizer) NormalizeDescription(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", validationerrors.ErrInvalidUTF8
	}

	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = norm.NFC.String(value)
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > n.descriptionLimit {
		return "", validationerrors.ErrTooLong
	}

	for _, r := range value {
		if !unicode.IsPrint(r) && r != '\t' && r != '\n' {
			return "", validationerrors.ErrDisallowedCharacter
		}
	}

	return value, nil
}

func (n *Normalizer) NormalizeSearch(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", validationerrors.ErrInvalidUTF8
	}

	value = norm.NFC.String(strings.TrimSpace(value))
	if utf8.RuneCountInString(value) > n.searchLimit {
		return "", validationerrors.ErrTooLong
	}

	for _, r := range value {
		if !unicode.IsPrint(r) {
			return "", validationerrors.ErrDisallowedCharacter
		}
	}

	return value, nil
}
