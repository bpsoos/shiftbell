package chores

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/profileselection"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

const (
	defaultBrowseOffset = 0
	defaultBrowseLimit  = 20
)

type browseQuery struct {
	Status    string
	OwnerMode string
	OwnerId   int
	Search    string
	Offset    int
	Limit     int
}

type browseRequest struct {
	hasMultipleProfiles bool
	params              choremodels.BrowseChoresParams
	responseURL         url.URL
}

func (h *Handler) browse(ctx *echo.Context) error {
	request, err := parseBrowseRequest(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
		)
	}

	if ctx.QueryParam("scope") == "mine" {
		if !h.profilesUIEnabled {
			return ctx.NoContent(http.StatusNotFound)
		}
		if !acceptsHTML(ctx) {
			return hypermedia.NotAcceptable(ctx)
		}
		return h.browseMine(ctx, request)
	}
	if h.profilesUIEnabled && acceptsHTML(ctx) {
		request.hasMultipleProfiles, err = h.profileService.HasMultipleProfiles(
			ctx.Request().Context(),
		)
		if err != nil {
			return h.renderProfileStateError(ctx, err)
		}
	}
	return h.browseCollection(ctx, request, false)
}

func (h *Handler) browseMine(ctx *echo.Context, request browseRequest) error {
	ctx.Response().Header().Set("Cache-Control", "private, no-store")
	ctx.Response().Header().Add(echo.HeaderVary, "Cookie")
	profiles, err := h.profileService.List(ctx.Request().Context())
	if err != nil {
		logging.Default().Error("resolve chore owner filter", "err", err)
		return h.renderError(ctx, http.StatusInternalServerError, apiErrorResponse{
			Error: "Could not load your profile. Try again.",
		})
	}
	if len(profiles) == 0 {
		return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/setup")
	}
	selected := profileselection.SelectedId(ctx.Request())
	if len(profiles) == 1 {
		selected = profiles[0].Id
	}
	for _, profile := range profiles {
		if profile.Id == selected {
			request.hasMultipleProfiles = len(profiles) > 1
			request.params.OwnerFilter = choremodels.OwnerFilter{
				Mode: choremodels.OwnerFilterProfile, Id: &profile.Id,
			}
			return h.browseCollection(ctx, request, true)
		}
	}
	return hypermedia.HTMLRedirect(
		ctx,
		http.StatusSeeOther,
		"/profiles/selection?"+url.Values{
			"return_to": {request.responseURL.RequestURI()},
		}.Encode(),
	)
}

func (h *Handler) browseCollection(
	ctx *echo.Context,
	request browseRequest,
	mine bool,
) error {
	page, err := h.service.Browse(
		ctx.Request().Context(),
		&request.params,
	)
	if err != nil {
		return h.renderBrowseError(ctx, err)
	}

	return h.renderCollection(
		ctx,
		http.StatusOK,
		newBrowseResponse(request, page),
		request.params,
		mine,
		request.hasMultipleProfiles,
		request.responseURL.Query().Has("search"),
		searchChanged(
			ctx.Request().Header.Get("HX-Current-URL"),
			request.responseURL.Query().Has("search"),
			request.params.Search,
		),
	)
}

func searchChanged(
	currentURLHeader string,
	searchOpen bool,
	search string,
) bool {
	if currentURLHeader == "" {
		return false
	}
	currentURL, err := url.Parse(currentURLHeader)
	if err != nil {
		return false
	}
	currentQuery := currentURL.Query()
	return currentURL.Path == choreCollectionHref &&
		(currentQuery.Has("search") != searchOpen ||
			currentQuery.Get("search") != search)
}

func parseBrowseRequest(ctx *echo.Context) (browseRequest, error) {
	query := browseQuery{Offset: defaultBrowseOffset, Limit: defaultBrowseLimit}
	err := echo.QueryParamsBinder(ctx).
		String("status", &query.Status).
		String("owner_mode", &query.OwnerMode).
		Int("owner_id", &query.OwnerId).
		String("search", &query.Search).
		Int("offset", &query.Offset).
		Int("limit", &query.Limit).
		BindError()
	if err != nil {
		return browseRequest{}, browseQueryError(err)
	}
	if len(ctx.QueryParams()["owner_mode"]) > 1 ||
		len(ctx.QueryParams()["owner_id"]) > 1 {
		return browseRequest{}, validationerrors.ErrInvalidOwnerFilter
	}
	ownerFilter := choremodels.OwnerFilter{
		Mode: choremodels.OwnerFilterMode(query.OwnerMode),
	}
	if ctx.QueryParams().Has("owner_id") {
		if ctx.QueryParam("owner_id") == "" {
			return browseRequest{}, validationerrors.ErrInvalidOwnerFilter
		}
		ownerFilter.Id = &query.OwnerId
	}
	return browseRequest{
		params: choremodels.BrowseChoresParams{
			Status:      choremodels.ChoreStatus(query.Status),
			OwnerFilter: ownerFilter,
			Search:      query.Search,
			Offset:      query.Offset,
			Limit:       query.Limit,
		},
		responseURL: *ctx.Request().URL,
	}, nil
}

func browseQueryError(err error) error {
	var bindingError *echo.BindingError
	if !errors.As(err, &bindingError) {
		return err
	}
	switch bindingError.Field {
	case "owner_id":
		return validationerrors.ErrInvalidOwnerFilter
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

func newBrowseResponse(
	request browseRequest,
	page *choremodels.ChorePage,
) choreCollectionResponse {
	items := make([]choreResponse, len(page.Chores))
	for i := range page.Chores {
		items[i] = newChoreResponse(&page.Chores[i])
	}
	return choreCollectionResponse{
		Items:   items,
		More:    page.More,
		Links:   browseLinks(request, page.More),
		Actions: api.Relations{createChoreNavigationAction()},
	}
}

func browseLinks(request browseRequest, more bool) api.Relations {
	links := api.Relations{{Rel: "self", Href: request.responseURL.RequestURI()}}
	if more {
		links = append(links, api.Relation{
			Rel: "next",
			Href: chorePageHref(
				&request.responseURL,
				request.params.Offset+request.params.Limit,
			),
		})
	}
	if request.params.Offset > 0 {
		links = append(links, api.Relation{
			Rel: "previous",
			Href: chorePageHref(
				&request.responseURL,
				max(0, request.params.Offset-request.params.Limit),
			),
		})
	}
	return links
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
		errors.Is(err, validationerrors.ErrInvalidOwnerFilter) ||
		errors.Is(err, validationerrors.ErrInvalidSearch) ||
		errors.Is(err, validationerrors.ErrInvalidOffset) ||
		errors.Is(err, validationerrors.ErrInvalidLimit)
}
