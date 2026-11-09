package usernamevalidation

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/bpsoos/shiftbell/internal/models/validation"
	"golang.org/x/text/unicode/norm"
)

type Config struct {
	MaxLength int
}

type Validator struct {
	maxLength int
}

func New(config Config) *Validator {
	return &Validator{maxLength: config.MaxLength}
}

func (v *Validator) Validate(username string) (string, error) {
	if !utf8.ValidString(username) {
		return "", validation.ErrInvalidUTF8
	}
	if username == "" {
		return "", validation.ErrRequired
	}

	length := 0
	for _, character := range username {
		length++
		if unicode.IsSpace(character) {
			return "", fmt.Errorf(
				"%w: whitespace %q (%U) at character %d is not allowed",
				validation.ErrDisallowedCharacter,
				character,
				character,
				length,
			)
		}
		if !unicode.IsPrint(character) {
			return "", fmt.Errorf(
				"%w: non-printable character %q (%U) at character %d is not allowed",
				validation.ErrDisallowedCharacter,
				character,
				character,
				length,
			)
		}
	}
	username = norm.NFC.String(username)
	length = utf8.RuneCountInString(username)
	if length > v.maxLength {
		return "", fmt.Errorf(
			"%w: maximum is %d characters; received %d",
			validation.ErrTooLong,
			v.maxLength,
			length,
		)
	}
	return username, nil
}
