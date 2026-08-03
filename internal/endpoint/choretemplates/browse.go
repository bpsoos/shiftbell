package choretemplates

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Browse(ctx *echo.Context) error {
	if !hypermedia.Accepts(ctx.Request()) {
		return ctx.NoContent(http.StatusNotAcceptable)
	}

	offset, err := strconv.Atoi(ctx.QueryParamOr("offset", "0"))
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: "invalid offset"},
		)
	}
	limit, err := strconv.Atoi(ctx.QueryParamOr("limit", "20"))
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: "invalid limit"},
		)
	}

	page, err := h.service.Browse(
		ctx.Request().Context(),
		&models.BrowseChoreTemplatesParams{
			Filter: models.ChoreTemplateFilter(ctx.QueryParamOr("state", "")),
			Search: ctx.QueryParamOr("search", ""),
			Offset: offset,
			Limit:  limit,
		},
	)
	if err != nil {
		logging.Default().Error("browse chore templates", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			errorResponse{Error: "something went wrong"},
		)
	}

	items := make([]response, len(page.ChoreTemplates))
	for i := range page.ChoreTemplates {
		items[i] = newResponse(&page.ChoreTemplates[i])
	}

	links := map[string]hypermedia.Link{
		"self": {Href: ctx.Request().URL.RequestURI()},
	}
	if page.More {
		links["next"] = hypermedia.Link{
			Href: choreTemplatePageHref(ctx.Request().URL, offset+limit),
		}
	}
	if offset > 0 {
		previousOffset := max(0, offset-limit)
		links["previous"] = hypermedia.Link{
			Href: choreTemplatePageHref(ctx.Request().URL, previousOffset),
		}
	}

	return hypermedia.JSON(ctx, http.StatusOK, collectionResponse{
		Items: items,
		More:  page.More,
		Links: links,
		Actions: map[string]hypermedia.Action{
			"create": createAction(),
		},
	})
}

func choreTemplatePageHref(requestURL *url.URL, offset int) string {
	pageURL := *requestURL
	query := pageURL.Query()
	query.Set("offset", strconv.Itoa(offset))
	pageURL.RawQuery = query.Encode()
	return pageURL.RequestURI()
}
