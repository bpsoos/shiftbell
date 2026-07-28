package chores

import "errors"

var (
	ErrNotFound                    = errors.New("chore not found")
	ErrActiveScheduledCannotDelete = errors.New("active scheduled chore cannot be deleted")
)
