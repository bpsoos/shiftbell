package schedules

import (
	"encoding/json"
	"mime"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

type editRequest struct {
	Name          string `json:"name"`
	IntervalDays  int    `json:"interval_days"`
	intervalValue string
}

func parseEditRequest(ctx *echo.Context) (editRequest, error) {
	contentType := ctx.Request().Header.Get(echo.HeaderContentType)
	if contentType == "" {
		return parseJSONEditRequest(ctx)
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return editRequest{}, binding.ErrUnsupportedMediaType
	}
	switch mediaType {
	case echo.MIMEApplicationForm:
		return parseFormEditRequest(ctx)
	case echo.MIMEApplicationJSON, hypermedia.MediaType:
		return parseJSONEditRequest(ctx)
	default:
		return editRequest{}, binding.ErrUnsupportedMediaType
	}
}

func parseJSONEditRequest(ctx *echo.Context) (editRequest, error) {
	var request editRequest
	err := json.NewDecoder(ctx.Request().Body).Decode(&request)
	request.intervalValue = strconv.Itoa(request.IntervalDays)
	return request, err
}

func parseFormEditRequest(ctx *echo.Context) (editRequest, error) {
	values, err := ctx.FormValues()
	if err != nil {
		return editRequest{}, err
	}
	request := editRequest{
		Name:          values.Get("name"),
		intervalValue: values.Get("interval_days"),
	}
	request.IntervalDays, err = strconv.Atoi(request.intervalValue)
	if err != nil {
		return request, validationerrors.ErrInvalidInterval
	}
	return request, nil
}
