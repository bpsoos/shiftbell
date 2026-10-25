package schedules

import (
	"strconv"
	"time"

	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
)

type response struct {
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

type representation struct {
	response
	Actions api.Relations `json:"_actions"`
}

type errorResponse struct {
	Error   string        `json:"error"`
	Links   api.Relations `json:"_links"`
	Actions api.Relations `json:"_actions,omitempty"`
}

func resourceHref(id int) string {
	return "/schedules/" + strconv.Itoa(id)
}

func collectionLink() api.Relations {
	return api.Relations{{Rel: "collection", Href: "/schedules"}}
}

func newRepresentation(schedule *models.Schedule) representation {
	actions := api.Relations{}
	if schedule.DeactivatedAt == nil {
		actions = append(
			actions,
			api.Relation{Rel: "edit", Href: resourceHref(schedule.Id)},
			api.Relation{
				Rel:  "deactivate",
				Href: resourceHref(schedule.Id) + "/deactivation",
			},
		)
	}
	return representation{response: newResponse(schedule), Actions: actions}
}

func newResponse(schedule *models.Schedule) response {
	links := api.Relations{
		{Rel: "self", Href: resourceHref(schedule.Id)},
		{Rel: "collection", Href: "/schedules"},
		{
			Rel:  "chore_template",
			Href: "/chore-templates/" + strconv.Itoa(schedule.ChoreTemplateId),
		},
	}
	if schedule.ActiveChoreId != nil {
		links = append(links, api.Relation{
			Rel: "active_chore", Href: "/chores/" + strconv.Itoa(*schedule.ActiveChoreId),
		})
	}
	return response{
		Id:                  schedule.Id,
		Name:                schedule.Name,
		ChoreTemplateId:     schedule.ChoreTemplateId,
		ChoreTemplateName:   schedule.ChoreTemplateName,
		IntervalDays:        schedule.IntervalDays,
		DeactivatedAt:       schedule.DeactivatedAt,
		ActiveChoreId:       schedule.ActiveChoreId,
		ActiveChoreDeadline: schedule.ActiveChoreDeadline,
		Links:               links,
	}
}
