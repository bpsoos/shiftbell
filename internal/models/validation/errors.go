package validation

import "errors"

var (
	ErrInvalidName         = errors.New("invalid name")
	ErrInvalidDescription  = errors.New("invalid description")
	ErrInvalidSearch       = errors.New("invalid search")
	ErrInvalidFilter       = errors.New("invalid filter")
	ErrInvalidOffset       = errors.New("invalid offset")
	ErrInvalidLimit        = errors.New("invalid limit")
	ErrInvalidInterval     = errors.New("invalid interval")
	ErrInvalidDeadline     = errors.New("invalid deadline")
	ErrInvalidUTF8         = errors.New("invalid UTF-8")
	ErrRequired            = errors.New("value is required")
	ErrTooLong             = errors.New("value is too long")
	ErrDisallowedCharacter = errors.New("value contains a disallowed character")
)
