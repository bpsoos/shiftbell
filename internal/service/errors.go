package service

import "errors"

var (
	ErrInvalidName        = errors.New("invalid name")
	ErrInvalidDescription = errors.New("invalid description")
	ErrInvalidSearch      = errors.New("invalid search")
	ErrInvalidFilter      = errors.New("invalid filter")
	ErrInvalidOffset      = errors.New("invalid offset")
	ErrInvalidLimit       = errors.New("invalid limit")
	ErrInvalidInterval    = errors.New("invalid interval")
)
