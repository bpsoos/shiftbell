package profiles

import "errors"

var (
	ErrNameConflict       = errors.New("a profile already uses this name")
	ErrNameContainsSpaces = errors.New("name contains spaces")
	ErrAlreadyInitialized = errors.New("a profile already exists")
)

type Profile struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type CreateParams struct {
	Name string
}
