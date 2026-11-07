package chores

import "errors"

var (
	ErrInvalidOwnerProfile = errors.New(
		"owner profile must be an existing profile",
	)
	ErrNotFound                    = errors.New("chore not found")
	ErrActiveScheduledCannotDelete = errors.New(
		"active scheduled chore cannot be deleted",
	)
)
