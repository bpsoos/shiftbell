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

type ActiveScheduleReference struct {
	Id   int
	Name string
}

type ActiveScheduleReferencesError struct {
	Schedules []ActiveScheduleReference
}

func (e *ActiveScheduleReferencesError) Error() string {
	return fmt.Sprintf("chore template has %d active schedule references", len(e.Schedules))
}
