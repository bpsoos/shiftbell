package choretemplates

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Browse(ctx *echo.Context) error {
	representationType := hypermedia.Negotiate(ctx.Request())
	if representationType == hypermedia.RepresentationUnsupported {
		return hypermedia.NotAcceptable(ctx)
	}

	offset, err := strconv.Atoi(ctx.QueryParamOr("offset", "0"))
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: "invalid offset"},
		)
	}
	limit, err := strconv.Atoi(ctx.QueryParamOr("limit", "20"))
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: "invalid limit"},
		)
	}

	pickerRequested := ctx.QueryParamOr("picker", "") == "1"
	filter := models.ChoreTemplateFilter(ctx.QueryParamOr("state", ""))
	search := ctx.QueryParamOr("search", "")
	responseURL := *ctx.Request().URL
	if representationType == hypermedia.RepresentationHTML {
		filter = models.ChoreTemplateFilterActive
		search = ""
		query := responseURL.Query()
		query.Del("search")
		query.Del("state")
		responseURL.RawQuery = query.Encode()
	} else if pickerRequested {
		filter = models.ChoreTemplateFilterActive
	}
	page, err := h.service.Browse(
		ctx.Request().Context(),
		&models.BrowseChoreTemplatesParams{
			Filter: filter,
			Search: search,
			Offset: offset,
			Limit:  limit,
		},
	)
	if err != nil {
		if isBrowseValidationError(err) {
			return h.renderError(
				ctx,
				http.StatusUnprocessableEntity,
				errorResponse{Error: err.Error(), Links: collectionLink()},
			)
		}
		logging.Default().Error("browse chore templates", "err", err)
		return h.renderError(
			ctx,
			http.StatusInternalServerError,
			errorResponse{Error: "something went wrong", Links: collectionLink()},
		)
	}

	links := api.Relations{
		{Rel: "self", Href: responseURL.RequestURI()},
	}
	if page.More {
		links = append(links, api.Relation{
			Rel:  "next",
			Href: choreTemplatePageHref(&responseURL, offset+limit),
		})
	}
	if offset > 0 {
		previousOffset := max(0, offset-limit)
		links = append(links, api.Relation{
			Rel:  "previous",
			Href: choreTemplatePageHref(&responseURL, previousOffset),
		})
	}
	if pickerRequested {
		items := make([]pickerItemResponse, len(page.ChoreTemplates))
		for i := range page.ChoreTemplates {
			items[i] = newPickerItemResponse(&page.ChoreTemplates[i])
		}
		return h.renderPicker(
			ctx,
			http.StatusOK,
			pickerCollectionResponse{
				Items: items,
				More:  page.More,
				Links: links,
			},
		)
	}

	items := make([]response, len(page.ChoreTemplates))
	for i := range page.ChoreTemplates {
		items[i] = newResponse(&page.ChoreTemplates[i])
	}

	return h.renderCollection(
		ctx,
		http.StatusOK,
		collectionResponse{
			Items:   items,
			More:    page.More,
			Links:   links,
			Actions: api.Relations{createAction()},
		},
	)
}

func isBrowseValidationError(err error) bool {
	return errors.Is(err, validationerrors.ErrInvalidFilter) ||
		errors.Is(err, validationerrors.ErrInvalidSearch) ||
		errors.Is(err, validationerrors.ErrInvalidOffset) ||
		errors.Is(err, validationerrors.ErrInvalidLimit)
}

func choreTemplatePageHref(requestURL *url.URL, offset int) string {
	pageURL := *requestURL
	query := pageURL.Query()
	query.Set("offset", strconv.Itoa(offset))
	pageURL.RawQuery = query.Encode()
	return pageURL.RequestURI()
}
