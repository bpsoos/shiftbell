package schedules

import (
	"time"

	api "github.com/bpsoos/shiftbell/internal/models/api"
)

type Response struct {
	Id                  int           `json:"id"`
	Name                string        `json:"name"`
	ChoreTemplateId     int           `json:"chore_template_id"`
	ChoreTemplateName   string        `json:"chore_template_name"`
	IntervalDays        int           `json:"interval_days"`
	DeactivatedAt       *time.Time    `json:"deactivated_at"`
	ActiveChoreId       *int          `json:"active_chore_id"`
	ActiveChoreDeadline *time.Time    `json:"active_chore_deadline"`
	Links               api.Relations `json:"_links"`
}

type Representation struct {
	Response
	Actions api.Relations `json:"_actions"`
}

type CollectionResponse struct {
	Items []Response    `json:"items"`
	More  bool          `json:"more"`
	Links api.Relations `json:"_links"`
}
