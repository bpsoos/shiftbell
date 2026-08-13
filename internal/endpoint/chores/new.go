package chores

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/labstack/echo/v5"
)

type (
	choreCreationChoice   = choreapimodels.CreationChoice
	choreCreationTemplate = choreapimodels.CreationTemplate
	choreCreationResponse = choreapimodels.CreationResponse
)

type newChoreQuery struct {
	TemplateID       int
	TemplateSelected bool
	Source           string
	Recurrence       string
}

func (h *Handler) newChore(ctx *echo.Context) error {
	query, err := parseNewChoreQuery(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
		)
	}
	if query.TemplateSelected {
		return h.newTemplateBasedChore(ctx, query)
	}
	return h.newManualChore(ctx, query)
}

func parseNewChoreQuery(ctx *echo.Context) (newChoreQuery, error) {
	query := newChoreQuery{
		TemplateSelected: ctx.QueryParam("template_id") != "",
	}
	err := echo.QueryParamsBinder(ctx).
		Int("template_id", &query.TemplateID).
		String("source", &query.Source).
		String("recurrence", &query.Recurrence).
		BindError()
	if err != nil || query.TemplateSelected && query.TemplateID <= 0 {
		return newChoreQuery{}, errInvalidChoreTemplateID
	}
	return query, nil
}

func (h *Handler) newManualChore(ctx *echo.Context, query newChoreQuery) error {
	if query.Source == "manual" && query.Recurrence == "scheduled" {
		return h.scheduledRecurrenceNotImplemented(ctx)
	}

	if query.Source == "manual" && query.Recurrence == "one-off" {
		action := createChoreSubmissionAction()
		return h.renderCreation(ctx, http.StatusOK, choreCreationResponse{
			Step:    "form",
			Actions: api.Relations{action},
		})
	}

	if query.Source == "manual" && query.Recurrence == "" {
		return h.renderCreation(ctx, http.StatusOK, choreCreationResponse{
			Step: "recurrence",
			Choices: []choreCreationChoice{
				{
					Label: "One-off",
					Href:  "/chores/new?source=manual&recurrence=one-off",
				},
				{
					Label: "Scheduled",
					Href:  "/chores/new?source=manual&recurrence=scheduled",
				},
			},
		})
	}

	return h.renderCreation(ctx, http.StatusOK, choreCreationResponse{
		Step: "source",
		Choices: []choreCreationChoice{
			{Label: "Specify new", Href: "/chores/new?source=manual"},
			{Label: "Select template", Href: "/chore-templates?picker=1"},
		},
	})
}

func (h *Handler) scheduledRecurrenceNotImplemented(ctx *echo.Context) error {
	return h.renderError(
		ctx,
		http.StatusNotImplemented,
		apiErrorResponse{Error: "scheduled recurrence is not implemented"},
	)
}

func (h *Handler) newTemplateBasedChore(
	ctx *echo.Context,
	query newChoreQuery,
) error {
	details, err := h.choreTemplateService.Get(
		ctx.Request().Context(),
		query.TemplateID,
	)
	if err != nil {
		return h.renderNewChoreTemplateError(ctx, err)
	}
	template := details.ChoreTemplate
	if template.DeactivatedAt != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: choretemplatemodels.ErrInactive.Error()},
		)
	}
	if query.Recurrence == "scheduled" {
		return h.scheduledRecurrenceNotImplemented(ctx)
	}
	if query.Recurrence == "" {
		return h.templateRecurrenceChoices(ctx, &template)
	}
	if query.Recurrence == "one-off" {
		return h.templateOneOffForm(ctx, &template)
	}
	return h.templateRecurrenceChoices(ctx, &template)
}

func (h *Handler) renderNewChoreTemplateError(
	ctx *echo.Context,
	err error,
) error {
	if errors.Is(err, choretemplatemodels.ErrNotFound) {
		return h.renderError(
			ctx,
			http.StatusNotFound,
			apiErrorResponse{Error: choretemplatemodels.ErrNotFound.Error()},
		)
	}
	logging.Default().Error("get chore template for chore creation", "err", err)
	return h.renderError(
		ctx,
		http.StatusInternalServerError,
		apiErrorResponse{Error: "something went wrong"},
	)
}

func (h *Handler) templateRecurrenceChoices(
	ctx *echo.Context,
	template *choretemplatemodels.ChoreTemplate,
) error {
	return h.renderCreation(ctx, http.StatusOK, choreCreationResponse{
		Step: "recurrence",
		Template: &choreCreationTemplate{
			Id:          template.Id,
			Name:        template.Name,
			Description: template.Description,
		},
		Choices: []choreCreationChoice{
			{
				Label: "One-off",
				Href: fmt.Sprintf(
					"/chores/new?template_id=%d&recurrence=one-off",
					template.Id,
				),
			},
			{
				Label: "Scheduled",
				Href: fmt.Sprintf(
					"/chores/new?template_id=%d&recurrence=scheduled",
					template.Id,
				),
			},
		},
	})
}

func (h *Handler) templateOneOffForm(
	ctx *echo.Context,
	template *choretemplatemodels.ChoreTemplate,
) error {
	action := createChoreSubmissionAction()
	return h.renderCreation(ctx, http.StatusOK, choreCreationResponse{
		Step: "form",
		Template: &choreCreationTemplate{
			Id:          template.Id,
			Name:        template.Name,
			Description: template.Description,
		},
		Actions: api.Relations{action},
	})
}
