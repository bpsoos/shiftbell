package choretemplates

import (
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choretemplateviewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
)

func (h *Handler) renderCollection(
	ctx *echo.Context,
	status int,
	representation collectionResponse,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, representation)
	case hypermedia.RepresentationHTML:
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Collection(
				collectionViewModel(representation),
				fullPage(ctx),
			),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderPicker(
	ctx *echo.Context,
	status int,
	representation pickerCollectionResponse,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, representation)
	case hypermedia.RepresentationHTML:
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Picker(
				pickerViewModel(representation),
				fullPage(ctx),
			),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderDetail(
	ctx *echo.Context,
	status int,
	representation representation,
	notice string,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, representation)
	case hypermedia.RepresentationHTML:
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Detail(detailViewModel(representation, notice), fullPage(ctx)),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderEditForm(
	ctx *echo.Context,
	status int,
	model choretemplateviewmodels.EditForm,
) error {
	return hypermedia.HTML(ctx, status, h.view.EditForm(model, fullPage(ctx)))
}

func (h *Handler) renderError(
	ctx *echo.Context,
	status int,
	representation errorResponse,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, representation)
	case hypermedia.RepresentationHTML:
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Error(errorViewModel(representation), fullPage(ctx)),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func fullPage(ctx *echo.Context) bool {
	return ctx.Request().Header.Get("HX-Request") != "true"
}

func supported(ctx *echo.Context) bool {
	return hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationUnsupported
}

func acceptsHTMXHTML(ctx *echo.Context) bool {
	return !fullPage(ctx) &&
		hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML
}
