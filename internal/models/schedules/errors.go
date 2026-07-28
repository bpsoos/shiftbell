package schedules

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("schedule not found")
	ErrInactive = errors.New("schedule inactive")
)

type NameConflictError struct {
	ExistingId int
}

func (e *NameConflictError) Error() string {
	return fmt.Sprintf("schedule name conflicts with active schedule %d", e.ExistingId)
}
