package schedules

import (
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
	"github.com/labstack/echo/v5"
)

func (h *Handler) EditForm(ctx *echo.Context) error {
	if hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationHTML {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseScheduleID(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: err.Error(), Links: collectionLink()},
		)
	}
	schedule, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderGetError(ctx, err)
	}
	if schedule.DeactivatedAt != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{
				Error: models.ErrInactive.Error(),
				Links: newRepresentation(schedule).Links,
			},
		)
	}
	return h.renderEditForm(ctx, http.StatusOK, viewmodels.EditForm{
		ActionHref: resourceHref(id), CancelHref: resourceHref(id),
		Name: schedule.Name, IntervalDays: strconv.Itoa(schedule.IntervalDays),
	})
}

func (h *Handler) renderEditForm(
	ctx *echo.Context,
	status int,
	model viewmodels.EditForm,
) error {
	return hypermedia.HTML(ctx, status, h.view.EditForm(model, fullPage(ctx)))
}
