package schedules

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/labstack/echo/v5"
)

type errorResponse struct {
	Error string        `json:"error"`
	Links api.Relations `json:"_links"`
}

func (h *Handler) Get(ctx *echo.Context) error {
	if hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationJSON {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseScheduleID(ctx)
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: err.Error(), Links: collectionLink()},
		)
	}
	schedule, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderGetError(ctx, err)
	}
	return hypermedia.JSON(ctx, http.StatusOK, newResponse(schedule))
}

func (h *Handler) renderGetError(ctx *echo.Context, err error) error {
	if errors.Is(err, models.ErrNotFound) {
		return hypermedia.JSON(
			ctx,
			http.StatusNotFound,
			errorResponse{Error: models.ErrNotFound.Error(), Links: collectionLink()},
		)
	}
	logging.Default().Error("get schedule", "err", err)
	return hypermedia.JSON(
		ctx,
		http.StatusInternalServerError,
		errorResponse{Error: "something went wrong", Links: collectionLink()},
	)
}

func collectionLink() api.Relations {
	return api.Relations{{Rel: "collection", Href: "/schedules"}}
}
