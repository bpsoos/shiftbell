package chores

import (
	"fmt"
	"net/http"
	"time"

	api "github.com/bpsoos/shiftbell/internal/models/api"
	choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
)

type (
	choreResponse           = choreapimodels.Response
	choreRepresentation     = choreapimodels.Representation
	choreCollectionResponse = choreapimodels.CollectionResponse
	apiErrorResponse        = api.ErrorResponse
)

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
		Links: map[string]api.Link{
			"self":       {Href: fmt.Sprintf("/chores/%d", chore.Id)},
			"collection": {Href: "/chores"},
		},
	}
}

func createChoreNavigationAction() api.Action {
	return api.Action{
		Href:   "/chores/new",
		Method: http.MethodGet,
	}
}

func createChoreSubmissionAction() api.Action {
	return api.Action{
		Href:        "/chores",
		Method:      http.MethodPost,
		ContentType: "application/json",
	}
}

func activeOneOffActions(selfHref string) map[string]api.Action {
	return map[string]api.Action{
		"edit": {
			Href:        selfHref,
			Method:      http.MethodPatch,
			ContentType: "application/json",
			Fields: []api.ActionField{
				{Name: "name", Type: "string", Required: true},
				{Name: "description", Type: "string", Required: false},
				{Name: "deadline", Type: "date", Required: true},
			},
		},
		"complete": {
			Href:        selfHref + "/completion",
			Method:      http.MethodPut,
			ContentType: "application/json",
			Fields: []api.ActionField{
				{Name: "completed_on", Type: "date", Required: true},
			},
		},
		"delete": {
			Href:   selfHref,
			Method: http.MethodDelete,
		},
	}
}

func completedOneOffActions(selfHref string) map[string]api.Action {
	return map[string]api.Action{
		"correct_completion": {
			Href:        selfHref + "/completion",
			Method:      http.MethodPatch,
			ContentType: "application/json",
			Fields: []api.ActionField{
				{Name: "completed_on", Type: "date", Required: true},
			},
		},
		"delete": {
			Href:   selfHref,
			Method: http.MethodDelete,
		},
	}
}
