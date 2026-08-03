package choretemplates

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("chore template not found")
	ErrInactive     = errors.New("chore template inactive")
	ErrNameConflict = errors.New(
		"chore template name conflicts with an active chore template",
	)
)

type ActiveScheduleReference struct {
	Id   int
	Name string
}

type ActiveScheduleReferencesError struct {
	Schedules []ActiveScheduleReference
}

func (e *ActiveScheduleReferencesError) Error() string {
	return fmt.Sprintf(
		"chore template has %d active schedule references",
		len(e.Schedules),
	)
}
