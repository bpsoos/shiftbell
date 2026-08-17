package chores

import (
	"fmt"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/routes"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
)

const choreCollectionHref = "/chores"

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
			{Rel: "self", Href: choreHref(chore.Id)},
			{Rel: "collection", Href: choreCollectionHref},
		},
	}
}

func newChoreRepresentation(chore *choremodels.Chore) choreRepresentation {
	return choreRepresentation{
		Response: newChoreResponse(chore),
		Actions:  actionsForChore(chore),
	}
}

func createChoreNavigationAction() api.Relation {
	return api.Relation{Rel: "create", Href: (routes.ChoreCreation{}).Href()}
}

func createChoreSubmissionAction() api.Relation {
	return api.Relation{Rel: "create", Href: choreCollectionHref}
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
	selfHref := choreHref(chore.Id)
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

func choreHref(id int) string {
	return fmt.Sprintf("%s/%d", choreCollectionHref, id)
}
