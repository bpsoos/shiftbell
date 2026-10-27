package schedules

import (
	"strconv"

	api "github.com/bpsoos/shiftbell/internal/models/api"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/schedules"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
)

type (
	response           = apimodels.Response
	representation     = apimodels.Representation
	collectionResponse = apimodels.CollectionResponse
)

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
	return representation{Response: newResponse(schedule), Actions: actions}
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
