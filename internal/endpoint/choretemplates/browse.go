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

const (
	defaultBrowseOffset = 0
	defaultBrowseLimit  = 20
)

type browseQuery struct {
	Filter string
	Search string
	Offset int
	Limit  int
	Picker string
}

type browseRequest struct {
	params          models.BrowseChoreTemplatesParams
	responseURL     url.URL
	pickerRequested bool
}

func (h *Handler) Browse(ctx *echo.Context) error {
	representationType := hypermedia.Negotiate(ctx.Request())
	if representationType == hypermedia.RepresentationUnsupported {
		return hypermedia.NotAcceptable(ctx)
	}

	request, err := parseBrowseRequest(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: err.Error()},
		)
	}
	page, err := h.service.Browse(
		ctx.Request().Context(),
		&request.params,
	)
	if err != nil {
		return h.renderBrowseError(ctx, err)
	}

	links := browseLinks(request, page.More)
	if request.pickerRequested {
		return h.renderPicker(ctx, http.StatusOK, pickerResponse(page, links))
	}
	return h.renderCollection(ctx, http.StatusOK, browseResponse(page, links))
}

func parseBrowseRequest(ctx *echo.Context) (browseRequest, error) {
	query := browseQuery{
		Offset: defaultBrowseOffset,
		Limit:  defaultBrowseLimit,
	}
	err := echo.QueryParamsBinder(ctx).
		String("state", &query.Filter).
		String("search", &query.Search).
		Int("offset", &query.Offset).
		Int("limit", &query.Limit).
		String("picker", &query.Picker).
		BindError()
	if err != nil {
		return browseRequest{}, browseQueryError(err)
	}

	request := browseRequest{
		params: models.BrowseChoreTemplatesParams{
			Filter: models.ChoreTemplateFilter(query.Filter),
			Search: query.Search,
			Offset: query.Offset,
			Limit:  query.Limit,
		},
		responseURL:     *ctx.Request().URL,
		pickerRequested: query.Picker == "1",
	}
	if request.pickerRequested {
		request.params.Filter = models.ChoreTemplateFilterActive
	}
	return request, nil
}

func browseQueryError(err error) error {
	var bindingError *echo.BindingError
	if !errors.As(err, &bindingError) {
		return err
	}
	switch bindingError.Field {
	case "offset":
		return validationerrors.ErrInvalidOffset
	case "limit":
		return validationerrors.ErrInvalidLimit
	default:
		return err
	}
}

func (h *Handler) renderBrowseError(ctx *echo.Context, err error) error {
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

func browseLinks(request browseRequest, more bool) api.Relations {
	links := api.Relations{{Rel: "self", Href: request.responseURL.RequestURI()}}
	if more {
		links = append(links, api.Relation{
			Rel: "next",
			Href: choreTemplatePageHref(
				&request.responseURL,
				request.params.Offset+request.params.Limit,
			),
		})
	}
	if request.params.Offset > 0 {
		links = append(links, api.Relation{
			Rel: "previous",
			Href: choreTemplatePageHref(
				&request.responseURL,
				max(0, request.params.Offset-request.params.Limit),
			),
		})
	}
	return links
}

func browseResponse(
	page *models.ChoreTemplatePage,
	links api.Relations,
) collectionResponse {
	items := make([]response, len(page.ChoreTemplates))
	for i := range page.ChoreTemplates {
		items[i] = newResponse(&page.ChoreTemplates[i])
	}
	return collectionResponse{
		Items:   items,
		More:    page.More,
		Links:   links,
		Actions: api.Relations{createAction()},
	}
}

func pickerResponse(
	page *models.ChoreTemplatePage,
	links api.Relations,
) pickerCollectionResponse {
	items := make([]pickerItemResponse, len(page.ChoreTemplates))
	for i := range page.ChoreTemplates {
		items[i] = newPickerItemResponse(&page.ChoreTemplates[i])
	}
	return pickerCollectionResponse{Items: items, More: page.More, Links: links}
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
