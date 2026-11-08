package profiles

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Delete(ctx *echo.Context) error {
	if !h.supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
			return hypermedia.HTML(
				ctx,
				http.StatusUnprocessableEntity,
				h.view.Error(viewmodels.Error{
					Message:        "Invalid profile. Return to profiles and try again.",
					CollectionHref: "/profiles",
				}, fullPage(ctx)),
			)
		}
		return hypermedia.JSON(ctx, http.StatusUnprocessableEntity, api.ErrorResponse{
			Error: "invalid profile id", Links: collectionLink(),
		})
	}
	if err := h.service.Delete(
		ctx.Request().Context(),
		id,
	); err != nil &&
		!errors.Is(err, models.ErrNotFound) {
		return h.internalError(ctx, "delete profile", err)
	}
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
		return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/profiles")
	}
	return hypermedia.NoContent(ctx, http.StatusNoContent)
}
