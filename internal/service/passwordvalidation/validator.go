package passwordvalidation

import (
	"fmt"
	"unicode/utf8"

	"github.com/bpsoos/shiftbell/internal/models/validation"
)

const minimumLength = 15

type Validator struct{}

func New() *Validator {
	return &Validator{}
}

func (*Validator) Validate(password string) error {
	if !utf8.ValidString(password) {
		return validation.ErrInvalidUTF8
	}
	if password == "" {
		return validation.ErrRequired
	}
	length := utf8.RuneCountInString(password)
	if length < minimumLength {
		return fmt.Errorf(
			"%w: password must contain at least %d characters; received %d",
			validation.ErrTooShort,
			minimumLength,
			length,
		)
	}
	return nil
}
