package chores

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

func (h *Handler) newChore(ctx *echo.Context) error {
	if templateId := ctx.QueryParamOr("template_id", ""); templateId != "" {
		return h.newTemplateBasedChore(ctx, templateId)
	}
	return h.newManualChore(ctx)
}

func (h *Handler) newManualChore(ctx *echo.Context) error {
	if ctx.QueryParamOr("source", "") == "manual" &&
		ctx.QueryParamOr("recurrence", "") == "scheduled" {
		return h.scheduledRecurrenceNotImplemented(ctx)
	}

	if ctx.QueryParamOr("source", "") == "manual" &&
		ctx.QueryParamOr("recurrence", "") == "one-off" {
		action := createChoreSubmissionAction()
		return h.renderCreation(ctx, http.StatusOK, choreCreationResponse{
			Step:    "form",
			Actions: api.Relations{action},
		})
	}

	if ctx.QueryParamOr("source", "") == "manual" &&
		ctx.QueryParamOr("recurrence", "") == "" {
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

func (h *Handler) newTemplateBasedChore(ctx *echo.Context, rawTemplateId string) error {
	templateId, err := strconv.Atoi(rawTemplateId)
	if err != nil || templateId <= 0 {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: "invalid chore template id"},
		)
	}
	details, err := h.choreTemplateService.Get(ctx.Request().Context(), templateId)
	if err != nil {
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
	template := details.ChoreTemplate
	if template.DeactivatedAt != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: choretemplatemodels.ErrInactive.Error()},
		)
	}
	recurrence := ctx.QueryParamOr("recurrence", "")
	if recurrence == "scheduled" {
		return h.scheduledRecurrenceNotImplemented(ctx)
	}
	if recurrence == "" {
		return h.templateRecurrenceChoices(ctx, &template)
	}
	if recurrence == "one-off" {
		return h.templateOneOffForm(ctx, &template)
	}
	return h.templateRecurrenceChoices(ctx, &template)
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
