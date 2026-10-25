package schedules

import "errors"

var (
	ErrNotFound     = errors.New("schedule not found")
	ErrInactive     = errors.New("schedule inactive")
	ErrNameConflict = errors.New("schedule name conflicts with an active schedule")
)
