package profiles

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/labstack/echo/v5"
)

func (h *Handler) CreateForm(ctx *echo.Context) error {
	if !h.supported(ctx) ||
		hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationHTML {
		return hypermedia.NotAcceptable(ctx)
	}
	return hypermedia.HTML(
		ctx,
		http.StatusOK,
		h.view.Form(newForm("", false), fullPage(ctx)),
	)
}
