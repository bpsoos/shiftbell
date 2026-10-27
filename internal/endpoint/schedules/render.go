package schedules

import (
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
	"github.com/labstack/echo/v5"
)

func (h *Handler) renderCollection(
	ctx *echo.Context,
	status int,
	collection collectionResponse,
	params models.BrowseSchedulesParams,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, collection)
	case hypermedia.RepresentationHTML:
		filter := params.Filter
		if filter == "" {
			filter = models.ScheduleFilterActive
		}
		return hypermedia.HTML(ctx, status, h.view.Collection(viewmodels.Collection{
			Collection: collection, Filter: filter, Search: params.Search,
		}, fullPage(ctx)))
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderDetail(
	ctx *echo.Context,
	status int,
	schedule representation,
	notice string,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, schedule)
	case hypermedia.RepresentationHTML:
		return hypermedia.HTML(
			ctx,
			status,
			h.view.Detail(
				viewmodels.Detail{Schedule: schedule, Notice: notice},
				fullPage(ctx),
			),
		)
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderError(
	ctx *echo.Context,
	status int,
	response errorResponse,
) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, status, response)
	case hypermedia.RepresentationHTML:
		return hypermedia.HTML(ctx, status, h.view.Error(viewmodels.Error{
			Message: response.Error, Links: response.Links, Actions: response.Actions,
		}, fullPage(ctx)))
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}

func (h *Handler) renderBrowseFeedback(
	ctx *echo.Context,
	status int,
	message string,
) error {
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationJSON {
		return hypermedia.JSON(ctx, status, map[string]string{"error": message})
	}
	return h.renderError(
		ctx,
		status,
		errorResponse{Error: message, Links: collectionLink()},
	)
}

func fullPage(ctx *echo.Context) bool {
	return ctx.Request().Header.Get("HX-Request") != "true"
}
