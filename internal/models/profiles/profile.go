package profiles

import "errors"

var (
	ErrNameConflict       = errors.New("a profile already uses this name")
	ErrAlreadyInitialized = errors.New("a profile already exists")
)

type Profile struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type CreateParams struct {
	Name string
}
