package schedules

import (
	api "github.com/bpsoos/shiftbell/internal/models/api"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/schedules"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
)

type Collection struct {
	Collection      apimodels.CollectionResponse
	Filter          models.ScheduleFilter
	Search          string
	SearchOpen      bool
	AutofocusSearch bool
}

type Detail struct {
	Schedule apimodels.Representation
	Notice   string
}

type Error struct {
	Message string
	Links   api.Relations
	Actions api.Relations
}

type EditForm struct {
	ActionHref        string
	CancelHref        string
	Name              string
	IntervalDays      string
	NameError         string
	IntervalDaysError string
	SummaryError      string
}
