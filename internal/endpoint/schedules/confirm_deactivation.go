package schedules

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	confirmationmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) ConfirmDeactivation(ctx *echo.Context) error {
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseScheduleID(ctx)
	if err != nil {
		return h.renderDeactivationDialog(
			ctx,
			http.StatusUnprocessableEntity,
			confirmationmodels.Dialog{
				Heading: "Unable to deactivate schedule", Error: err.Error(),
			},
		)
	}
	return h.loadDeactivationDialog(ctx, id, http.StatusOK, "")
}

func (h *Handler) loadDeactivationDialog(
	ctx *echo.Context,
	id int,
	status int,
	message string,
) error {
	schedule, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		status = http.StatusInternalServerError
		if errors.Is(err, models.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			logging.Default().Error("load schedule for deactivation", "err", err)
		}
		return h.renderDeactivationDialog(ctx, status, confirmationmodels.Dialog{
			Heading: "Unable to deactivate schedule",
			Error:   "The schedule could not be loaded. Close this dialog and try again.",
		})
	}
	if schedule.DeactivatedAt != nil {
		return h.renderDeactivationDialog(
			ctx,
			http.StatusUnprocessableEntity,
			confirmationmodels.Dialog{
				Heading:    "Schedule already deactivated",
				Supporting: []string{"This schedule is permanently deactivated."},
			},
		)
	}
	return h.renderDeactivationDialog(ctx, status, deactivationDialog(schedule, message))
}

func deactivationDialog(
	schedule *models.Schedule,
	message string,
) confirmationmodels.Dialog {
	return confirmationmodels.Dialog{
		Heading: "Deactivate schedule?", Prompt: "Deactivate", Name: schedule.Name,
		ActionHref: resourceHref(schedule.Id) + "/deactivation", ActionMethod: "put",
		ActionLabel: "Deactivate permanently", Icon: "archive", Error: message,
	}
}

func (h *Handler) renderDeactivationDialog(
	ctx *echo.Context,
	status int,
	model confirmationmodels.Dialog,
) error {
	return hypermedia.HTML(ctx, status, h.view.ConfirmationDialog(model))
}
