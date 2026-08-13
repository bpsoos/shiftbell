package choretemplates

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) ConfirmDeactivation(ctx *echo.Context) error {
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}

	id, err := parseChoreTemplateID(ctx)
	if err != nil {
		return hypermedia.NoContent(ctx, http.StatusUnprocessableEntity)
	}
	details, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderDeactivationLoadError(ctx, err)
	}
	if details.DeactivatedAt == nil {
		return h.renderConfirmationDialog(
			ctx,
			http.StatusOK,
			deactivationDialogViewModel(&details.ChoreTemplate, ""),
		)
	}
	return h.renderConfirmationDialog(
		ctx,
		http.StatusUnprocessableEntity,
		inactiveDeactivationDialogViewModel(&details.ChoreTemplate),
	)
}

func inactiveDeactivationDialogViewModel(
	choreTemplate *models.ChoreTemplate,
) confirmationviewmodels.Dialog {
	return confirmationviewmodels.Dialog{
		Heading:    "Template already deactivated",
		Supporting: []string{"This template is no longer available for active use."},
		Error:      models.ErrInactive.Error(),
	}
}

func deactivationDialogViewModel(
	choreTemplate *models.ChoreTemplate,
	errorMessage string,
) confirmationviewmodels.Dialog {
	return confirmationviewmodels.Dialog{
		Heading: "Deactivate template?",
		Prompt:  "Deactivate",
		Name:    choreTemplate.Name,
		Supporting: []string{
			"It will no longer appear in template selectors.",
			"Existing chores are not changed.",
			"This cannot be reversed.",
		},
		ActionHref:   newRepresentation(choreTemplate).Actions.Href("deactivate"),
		ActionMethod: "put",
		ActionLabel:  "Deactivate permanently",
		Error:        errorMessage,
		Icon:         "archive",
	}
}

func (h *Handler) renderDeactivationLoadError(ctx *echo.Context, err error) error {
	if errors.Is(err, models.ErrNotFound) {
		return hypermedia.NoContent(ctx, http.StatusNotFound)
	}
	logging.Default().Error("load chore template for deactivation", "err", err)
	return hypermedia.NoContent(ctx, http.StatusInternalServerError)
}
