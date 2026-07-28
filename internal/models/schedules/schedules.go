package schedules

import "time"

type ScheduleFilter string

const (
	ScheduleFilterActive      ScheduleFilter = "active"
	ScheduleFilterDeactivated ScheduleFilter = "deactivated"
)

type BrowseSchedulesParams struct {
	Filter ScheduleFilter
	Search string
	Offset int
	Limit  int
}

type Schedule struct {
	Id                  int
	Name                string
	ChoreTemplateId     int
	ChoreTemplateName   string
	IntervalDays        int
	DeactivatedAt       *time.Time
	ActiveChoreId       *int
	ActiveChoreDeadline *time.Time
}

type ScheduleDetails = Schedule

type SchedulePage struct {
	Schedules []Schedule
	More      bool
}

type EditScheduleParams struct {
	Id           int
	Name         string
	IntervalDays int
}

type DeactivateScheduleParams struct {
	Id            int
	DeactivatedAt time.Time
}
