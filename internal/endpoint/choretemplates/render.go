package choretemplates

import (
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choretemplateviewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) renderCollection(
	ctx *echo.Context,
	status int,
	representation collectionResponse,
	filter choretemplatemodels.ChoreTemplateFilter,
	search string,
	searchOpen bool,
	autofocusSearch bool,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, representation)
	case hypermedia.RepresentationHTML:
		ctx.Response().Header().Add(echo.HeaderVary, "HX-Current-URL")
		model := collectionViewModel(
			representation,
			filter,
			search,
			searchOpen,
			autofocusSearch,
		)
		model.Notice = consumeFlashCookie(ctx)
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Collection(model, fullPage(ctx)),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderPicker(
	ctx *echo.Context,
	status int,
	representation pickerCollectionResponse,
	search string,
	autofocusSearch bool,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, representation)
	case hypermedia.RepresentationHTML:
		ctx.Response().Header().Add(echo.HeaderVary, "HX-Current-URL")
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Picker(
				pickerViewModel(representation, search, autofocusSearch),
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

func (h *Handler) renderConfirmationDialog(
	ctx *echo.Context,
	status int,
	model confirmationviewmodels.Dialog,
) error {
	return hypermedia.HTML(ctx, status, h.view.ConfirmationDialog(model))
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
