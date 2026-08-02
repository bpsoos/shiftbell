package chores

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
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
