package choretemplates

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
)

func (h *Handler) CreateForm(ctx *echo.Context) error {
	if !acceptsHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.renderCreateForm(ctx, http.StatusOK, viewmodels.CreateForm{
		ActionHref: choreTemplateCollectionHref, CancelHref: choreTemplateCollectionHref,
	})
}

func (h *Handler) renderCreateForm(
	ctx *echo.Context,
	status int,
	form viewmodels.CreateForm,
) error {
	return hypermedia.HTML(ctx, status, h.view.CreateForm(form, fullPage(ctx)))
}
