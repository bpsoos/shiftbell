package chores

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) renderCollection(
	ctx *echo.Context,
	status int,
	collection choreCollectionResponse,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, collection)
	case hypermedia.RepresentationHTML:
		model := collectionViewModel(collection)
		model.Notice = consumeFlashCookie(ctx)
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Collection(
				model,
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
	chore choreRepresentation,
	notice string,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, chore)
	case hypermedia.RepresentationHTML:
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Detail(detailViewModel(chore, notice), fullPage(ctx)),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderEditForm(
	ctx *echo.Context,
	status int,
	model choreviewmodels.EditForm,
) error {
	return hypermedia.HTML(
		ctx,
		status,
		h.view.EditForm(model, fullPage(ctx)),
	)
}

func (h *Handler) renderCompletionDialog(
	ctx *echo.Context,
	status int,
	model choreviewmodels.CompletionDialog,
) error {
	return hypermedia.HTML(ctx, status, h.view.CompletionDialog(model))
}

func (h *Handler) renderCreation(
	ctx *echo.Context,
	status int,
	creation choreCreationResponse,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, creation)
	case hypermedia.RepresentationHTML:
		if creation.Step == "form" && creation.Template == nil {
			return hypermedia.HTML(
				ctx,
				status,
				h.view.ManualOneOffForm(
					manualOneOffFormViewModel(creation),
					fullPage(ctx),
				),
			)
		}
		if creation.Step == "form" {
			return hypermedia.HTML(
				ctx,
				status,
				h.view.TemplateOneOffForm(
					templateOneOffFormViewModel(creation),
					fullPage(ctx),
				),
			)
		}
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Creation(creationViewModel(creation), fullPage(ctx)),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderError(
	ctx *echo.Context,
	status int,
	response apiErrorResponse,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, response)
	case hypermedia.RepresentationHTML:
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Error(errorViewModel(response), fullPage(ctx)),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderManualOneOffFormError(
	ctx *echo.Context,
	status int,
	feedback formFeedback,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, feedback.Error)
	case hypermedia.RepresentationHTML:
		return hypermedia.HTML(
			ctx,
			status,
			h.view.ManualOneOffForm(
				manualOneOffFormErrorViewModel(feedback),
				fullPage(ctx),
			),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderCreateFormError(
	ctx *echo.Context,
	status int,
	request createChoreRequest,
	feedback formFeedback,
) error {
	feedback.Action = createChoreSubmissionAction()
	feedback.CancelHref = choreCollectionHref
	if request.ChoreTemplateId != nil {
		return h.renderTemplateOneOffFormError(ctx, status, feedback)
	}
	return h.renderManualOneOffFormError(ctx, status, feedback)
}

func (h *Handler) renderTemplateOneOffFormError(
	ctx *echo.Context,
	status int,
	feedback formFeedback,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, feedback.Error)
	case hypermedia.RepresentationHTML:
		return hypermedia.HTML(
			ctx,
			status,
			h.view.TemplateOneOffForm(
				templateOneOffFormErrorViewModel(feedback),
				fullPage(ctx),
			),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderCreated(
	ctx *echo.Context,
	chore choreRepresentation,
) error {
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
		return renderHTMLRedirect(ctx, "/chores")
	}
	ctx.Response().Header().Set(
		echo.HeaderLocation,
		chore.Links.Href("self"),
	)
	return hypermedia.JSON(ctx, http.StatusCreated, chore)
}

func renderHTMLRedirect(ctx *echo.Context, href string) error {
	return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, href)
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
