package validation

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidUsername       = errors.New("invalid username")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrInvalidName           = errors.New("invalid name")
	ErrInvalidScheduleName   = fmt.Errorf("%w", ErrInvalidName)
	ErrInvalidDescription    = errors.New("invalid description")
	ErrInvalidSearch         = errors.New("invalid search")
	ErrInvalidFilter         = errors.New("invalid filter")
	ErrInvalidOwnerFilter    = errors.New("invalid owner filter")
	ErrInvalidOffset         = errors.New("invalid offset")
	ErrInvalidLimit          = errors.New("invalid limit")
	ErrInvalidInterval       = errors.New("invalid interval")
	ErrInvalidDeadline       = errors.New("invalid deadline")
	ErrInvalidCompletionDate = errors.New("invalid completion date")
	ErrInvalidUTF8           = errors.New("invalid UTF-8")
	ErrRequired              = errors.New("value is required")
	ErrTooShort              = errors.New("value is too short")
	ErrTooLong               = errors.New("value is too long")
	ErrDisallowedCharacter   = errors.New("value contains a disallowed character")
)
