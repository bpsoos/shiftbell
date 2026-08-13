package chores

import (
	"fmt"
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
		Links: api.Relations{
			{Rel: "self", Href: fmt.Sprintf("/chores/%d", chore.Id)},
			{Rel: "collection", Href: "/chores"},
		},
	}
}

func createChoreNavigationAction() api.Relation {
	return api.Relation{Rel: "create", Href: "/chores/new"}
}

func createChoreSubmissionAction() api.Relation {
	return api.Relation{Rel: "create", Href: "/chores"}
}

func activeOneOffActions(selfHref string) api.Relations {
	return api.Relations{
		{Rel: "edit", Href: selfHref},
		{Rel: "complete", Href: selfHref + "/completion"},
		{Rel: "delete", Href: selfHref},
	}
}

func activeScheduledActions(selfHref string) api.Relations {
	return api.Relations{
		{Rel: "edit", Href: selfHref},
		{Rel: "complete", Href: selfHref + "/completion"},
	}
}

func completedChoreActions(selfHref string) api.Relations {
	return api.Relations{
		{Rel: "correct_completion", Href: selfHref + "/completion"},
		{Rel: "delete", Href: selfHref},
	}
}

func actionsForChore(chore *choremodels.Chore) api.Relations {
	selfHref := fmt.Sprintf("/chores/%d", chore.Id)
	switch chore.Status {
	case choremodels.ChoreStatusActive:
		if chore.ScheduleId != nil {
			return activeScheduledActions(selfHref)
		}
		return activeOneOffActions(selfHref)
	case choremodels.ChoreStatusCompleted:
		return completedChoreActions(selfHref)
	default:
		return api.Relations{}
	}
}
