package hypermedia

import (
	"encoding/json"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"
)

const MediaType = "application/vnd.shiftbell+json"

type Representation uint8

const (
	RepresentationUnsupported Representation = iota
	RepresentationJSON
	RepresentationHTML
)

func Accepts(request *http.Request) bool {
	return Negotiate(request) == RepresentationJSON
}

func Negotiate(request *http.Request) Representation {
	accept := strings.Join(request.Header.Values("Accept"), ",")
	if accept == "" {
		return RepresentationHTML
	}

	ranges := parseAccept(accept)
	jsonPreference := selectPreference(MediaType, ranges)
	htmlPreference := selectPreference(echo.MIMETextHTML, ranges)

	if !jsonPreference.acceptable() && !htmlPreference.acceptable() {
		return RepresentationUnsupported
	}
	if !htmlPreference.acceptable() {
		return RepresentationJSON
	}
	if !jsonPreference.acceptable() {
		return RepresentationHTML
	}
	if jsonPreference.quality > htmlPreference.quality {
		return RepresentationJSON
	}
	if htmlPreference.quality > jsonPreference.quality {
		return RepresentationHTML
	}
	if jsonPreference.specificity > htmlPreference.specificity {
		return RepresentationJSON
	}
	if htmlPreference.specificity > jsonPreference.specificity {
		return RepresentationHTML
	}
	if jsonPreference.order < htmlPreference.order {
		return RepresentationJSON
	}

	return RepresentationHTML
}

type acceptRange struct {
	mediaType string
	quality   float64
	order     int
}

type preference struct {
	quality     float64
	specificity int
	order       int
	matched     bool
}

func (p preference) acceptable() bool {
	return p.matched && p.quality > 0
}

func parseAccept(value string) []acceptRange {
	ranges := make([]acceptRange, 0)
	order := 0
	for item := range strings.SplitSeq(value, ",") {
		mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(item))
		if err != nil {
			order++
			continue
		}
		quality := 1.0
		if rawQuality, ok := params["q"]; ok {
			quality, err = strconv.ParseFloat(rawQuality, 64)
			if err != nil || quality < 0 || quality > 1 {
				order++
				continue
			}
		}
		ranges = append(ranges, acceptRange{
			mediaType: mediaType,
			quality:   quality,
			order:     order,
		})
		order++
	}
	return ranges
}

func selectPreference(mediaType string, ranges []acceptRange) preference {
	selected := preference{}
	for _, candidate := range ranges {
		specificity := matchSpecificity(mediaType, candidate.mediaType)
		if specificity < 0 {
			continue
		}
		if !selected.matched || specificity > selected.specificity ||
			(specificity == selected.specificity && candidate.quality > selected.quality) {
			selected = preference{
				quality:     candidate.quality,
				specificity: specificity,
				order:       candidate.order,
				matched:     true,
			}
		}
	}
	return selected
}

func matchSpecificity(mediaType string, mediaRange string) int {
	if mediaRange == "*/*" {
		return 0
	}
	mediaRangeType, mediaRangeSubtype, ok := strings.Cut(mediaRange, "/")
	if !ok {
		return -1
	}
	mediaTypeType, _, ok := strings.Cut(mediaType, "/")
	if !ok || mediaRangeType != mediaTypeType {
		return -1
	}
	if mediaRangeSubtype == "*" {
		return 1
	}
	if mediaRange == mediaType {
		return 2
	}
	return -1
}

func JSON(ctx *echo.Context, status int, value any) error {
	ctx.Response().Header().Set(echo.HeaderContentType, MediaType)
	varyOn(ctx.Response().Header(), echo.HeaderAccept)
	ctx.Response().WriteHeader(status)
	return json.NewEncoder(ctx.Response()).Encode(value)
}

func HTML(ctx *echo.Context, status int, component templ.Component) error {
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	varyOn(ctx.Response().Header(), echo.HeaderAccept, "HX-Request")
	ctx.Response().WriteHeader(status)
	return component.Render(ctx.Request().Context(), ctx.Response())
}

func NotAcceptable(ctx *echo.Context) error {
	varyOn(ctx.Response().Header(), echo.HeaderAccept)
	return ctx.NoContent(http.StatusNotAcceptable)
}

func NoContent(ctx *echo.Context, status int) error {
	varyOn(ctx.Response().Header(), echo.HeaderAccept)
	return ctx.NoContent(status)
}

func HTMLRedirect(ctx *echo.Context, status int, location string) error {
	varyOn(ctx.Response().Header(), echo.HeaderAccept, "HX-Request")
	if ctx.Request().Header.Get("HX-Request") == "true" {
		ctx.Response().Header().Set("HX-Redirect", location)
		return ctx.NoContent(http.StatusOK)
	}
	return ctx.Redirect(status, location)
}

func varyOn(header http.Header, fields ...string) {
	values := make([]string, 0)
	for value := range strings.SplitSeq(strings.Join(header.Values(echo.HeaderVary), ","), ",") {
		value = strings.TrimSpace(value)
		if value != "" {
			values = append(values, value)
		}
	}
	for _, field := range fields {
		found := false
		for _, value := range values {
			if strings.EqualFold(value, field) {
				found = true
				break
			}
		}
		if !found {
			values = append(values, field)
		}
	}
	header.Set(echo.HeaderVary, strings.Join(values, ", "))
}
