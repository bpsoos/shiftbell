package choretemplates

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("chore template not found")
	ErrInactive = errors.New("chore template inactive")
)

type NameConflictError struct {
	ExistingId int
}

func (e *NameConflictError) Error() string {
	return fmt.Sprintf("chore template name conflicts with active chore template %d", e.ExistingId)
}
