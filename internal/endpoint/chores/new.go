package chores

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/labstack/echo/v5"
)

type choreCreationChoice struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type choreCreationTemplate struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type choreCreationResponse struct {
	Step     string                   `json:"step"`
	Template *choreCreationTemplate   `json:"template,omitempty"`
	Choices  []choreCreationChoice    `json:"choices,omitempty"`
	Fields   []hypermedia.ActionField `json:"fields,omitempty"`
	Action   *hypermedia.Action       `json:"action,omitempty"`
}

func (h *Handler) newChore(ctx *echo.Context) error {
	if templateId := ctx.QueryParamOr("template_id", ""); templateId != "" {
		return h.newTemplateBasedChore(ctx, templateId)
	}
	return h.newManualChore(ctx)
}

func (h *Handler) newManualChore(ctx *echo.Context) error {
	if ctx.QueryParamOr("source", "") == "manual" &&
		ctx.QueryParamOr("recurrence", "") == "scheduled" {
		return scheduledRecurrenceNotImplemented(ctx)
	}

	if ctx.QueryParamOr("source", "") == "manual" &&
		ctx.QueryParamOr("recurrence", "") == "one-off" {
		return hypermedia.JSON(ctx, http.StatusOK, choreCreationResponse{
			Step: "form",
			Fields: []hypermedia.ActionField{
				{Name: "name", Type: "string", Required: true},
				{Name: "description", Type: "string", Required: false},
				{Name: "deadline", Type: "date", Required: true},
				{Name: "save_as_chore_template", Type: "boolean", Required: false},
			},
			Action: &hypermedia.Action{
				Href:        "/chores",
				Method:      http.MethodPost,
				ContentType: "application/json",
			},
		})
	}

	if ctx.QueryParamOr("source", "") == "manual" &&
		ctx.QueryParamOr("recurrence", "") == "" {
		return hypermedia.JSON(ctx, http.StatusOK, choreCreationResponse{
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

	return hypermedia.JSON(ctx, http.StatusOK, choreCreationResponse{
		Step: "source",
		Choices: []choreCreationChoice{
			{Label: "Specify new", Href: "/chores/new?source=manual"},
			{Label: "Select template", Href: "/chore-templates?picker=1"},
		},
	})
}

func scheduledRecurrenceNotImplemented(ctx *echo.Context) error {
	return hypermedia.JSON(
		ctx,
		http.StatusNotImplemented,
		apiErrorResponse{Error: "scheduled recurrence is not implemented"},
	)
}

func (h *Handler) newTemplateBasedChore(ctx *echo.Context, rawTemplateId string) error {
	templateId, err := strconv.Atoi(rawTemplateId)
	if err != nil || templateId <= 0 {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: "invalid chore template id"},
		)
	}
	details, err := h.choreTemplatePersister.Get(ctx.Request().Context(), templateId)
	if err != nil {
		if errors.Is(err, choretemplatemodels.ErrNotFound) {
			return hypermedia.JSON(
				ctx,
				http.StatusNotFound,
				apiErrorResponse{Error: choretemplatemodels.ErrNotFound.Error()},
			)
		}
		logging.Default().Error("get chore template for chore creation", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}
	template := details.ChoreTemplate
	if template.DeactivatedAt != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: choretemplatemodels.ErrInactive.Error()},
		)
	}
	recurrence := ctx.QueryParamOr("recurrence", "")
	if recurrence == "scheduled" {
		return scheduledRecurrenceNotImplemented(ctx)
	}
	if recurrence == "" {
		return templateRecurrenceChoices(ctx, &template)
	}
	if recurrence == "one-off" {
		return templateOneOffForm(ctx, &template)
	}
	return templateRecurrenceChoices(ctx, &template)
}

func templateRecurrenceChoices(
	ctx *echo.Context,
	template *choretemplatemodels.ChoreTemplate,
) error {
	return hypermedia.JSON(ctx, http.StatusOK, choreCreationResponse{
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

func templateOneOffForm(
	ctx *echo.Context,
	template *choretemplatemodels.ChoreTemplate,
) error {
	return hypermedia.JSON(ctx, http.StatusOK, choreCreationResponse{
		Step: "form",
		Template: &choreCreationTemplate{
			Id:   template.Id,
			Name: template.Name,
		},
		Fields: []hypermedia.ActionField{
			{Name: "deadline", Type: "date", Required: true},
		},
		Action: &hypermedia.Action{
			Href:        "/chores",
			Method:      http.MethodPost,
			ContentType: "application/json",
		},
	})
}
