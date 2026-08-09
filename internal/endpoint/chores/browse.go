package chores

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) browse(ctx *echo.Context) error {
	offset, err := strconv.Atoi(ctx.QueryParamOr("offset", "0"))
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: "invalid offset"},
		)
	}
	limit, err := strconv.Atoi(ctx.QueryParamOr("limit", "20"))
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: "invalid limit"},
		)
	}

	selectedStatus := choremodels.ChoreStatusActive
	search := ""
	collectionURL := *ctx.Request().URL
	if hypermedia.Accepts(ctx.Request()) {
		selectedStatus = choremodels.ChoreStatus(ctx.QueryParamOr("status", "active"))
		search = ctx.QueryParamOr("search", "")
	} else {
		query := collectionURL.Query()
		query.Del("status")
		query.Del("search")
		collectionURL.RawQuery = query.Encode()
	}
	page, err := h.service.Browse(
		ctx.Request().Context(),
		&choremodels.BrowseChoresParams{
			Status: selectedStatus,
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
				apiErrorResponse{Error: err.Error()},
			)
		}
		logging.Default().Error("browse chores", "err", err)
		return h.renderError(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}

	items := make([]choreResponse, len(page.Chores))
	for i := range page.Chores {
		items[i] = newChoreResponse(&page.Chores[i])
	}

	links := api.Relations{
		{Rel: "self", Href: collectionURL.RequestURI()},
	}
	if page.More {
		links = append(links, api.Relation{
			Rel:  "next",
			Href: chorePageHref(&collectionURL, offset+limit),
		})
	}
	if offset > 0 {
		links = append(links, api.Relation{
			Rel:  "previous",
			Href: chorePageHref(&collectionURL, max(0, offset-limit)),
		})
	}

	return h.renderCollection(ctx, http.StatusOK, choreCollectionResponse{
		Items:   items,
		More:    page.More,
		Links:   links,
		Actions: api.Relations{createChoreNavigationAction()},
	})
}

func chorePageHref(requestURL *url.URL, offset int) string {
	pageURL := *requestURL
	query := pageURL.Query()
	query.Set("offset", strconv.Itoa(offset))
	pageURL.RawQuery = query.Encode()
	return pageURL.RequestURI()
}

func isBrowseValidationError(err error) bool {
	return errors.Is(err, validationerrors.ErrInvalidFilter) ||
		errors.Is(err, validationerrors.ErrInvalidSearch) ||
		errors.Is(err, validationerrors.ErrInvalidOffset) ||
		errors.Is(err, validationerrors.ErrInvalidLimit)
}
