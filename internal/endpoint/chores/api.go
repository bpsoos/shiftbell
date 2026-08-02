package chores

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
)

type choreResponse struct {
	Id          int                        `json:"id"`
	ScheduleId  *int                       `json:"schedule_id"`
	Status      choremodels.ChoreStatus    `json:"status"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Deadline    string                     `json:"deadline"`
	CompletedOn *string                    `json:"completed_on"`
	Links       map[string]hypermedia.Link `json:"_links"`
}

type choreRepresentation struct {
	choreResponse
	Actions map[string]hypermedia.Action `json:"_actions"`
}

type choreCollectionResponse struct {
	Items   []choreResponse              `json:"items"`
	More    bool                         `json:"more"`
	Links   map[string]hypermedia.Link   `json:"_links"`
	Actions map[string]hypermedia.Action `json:"_actions"`
}

type apiErrorResponse struct {
	Error string `json:"error"`
}

func newChoreResponse(chore *choremodels.Chore) choreResponse {
	var completedOn *string
	if !chore.CompletedOn.IsZero() {
		value := chore.CompletedOn.Format(time.DateOnly)
		completedOn = &value
	}

	return choreResponse{
		Id:          chore.Id,
		ScheduleId:  chore.ScheduleId,
		Status:      chore.Status,
		Name:        chore.Name,
		Description: chore.Description,
		Deadline:    chore.Deadline.Format(time.DateOnly),
		CompletedOn: completedOn,
		Links: map[string]hypermedia.Link{
			"self":       {Href: fmt.Sprintf("/chores/%d", chore.Id)},
			"collection": {Href: "/chores"},
		},
	}
}

func createChoreNavigationAction() hypermedia.Action {
	return hypermedia.Action{
		Href:   "/chores/new",
		Method: http.MethodGet,
	}
}

func activeOneOffActions(selfHref string) map[string]hypermedia.Action {
	return map[string]hypermedia.Action{
		"edit": {
			Href:        selfHref,
			Method:      http.MethodPatch,
			ContentType: "application/json",
			Fields: []hypermedia.ActionField{
				{Name: "name", Type: "string", Required: true},
				{Name: "description", Type: "string", Required: false},
				{Name: "deadline", Type: "date", Required: true},
			},
		},
		"complete": {
			Href:        selfHref + "/completion",
			Method:      http.MethodPut,
			ContentType: "application/json",
			Fields: []hypermedia.ActionField{
				{Name: "completed_on", Type: "date", Required: true},
			},
		},
		"delete": {
			Href:   selfHref,
			Method: http.MethodDelete,
		},
	}
}
