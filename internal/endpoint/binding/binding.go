package binding

import (
	"encoding/json"
	"errors"
	"io"
	"mime"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/labstack/echo/v5"
)

var ErrUnsupportedMediaType = errors.New("unsupported media type")

func Bind(ctx *echo.Context, target any) error {
	mediaType, _, err := mime.ParseMediaType(
		ctx.Request().Header.Get(echo.HeaderContentType),
	)
	if err != nil {
		return ErrUnsupportedMediaType
	}

	switch mediaType {
	case echo.MIMEApplicationJSON, hypermedia.MediaType:
		decoder := json.NewDecoder(ctx.Request().Body)
		if err := decoder.Decode(target); err != nil {
			return err
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			if err == nil {
				return errors.New("multiple JSON values")
			}
			return err
		}
		return nil
	case echo.MIMEApplicationForm, echo.MIMEMultipartForm:
		return ctx.Bind(target)
	default:
		return ErrUnsupportedMediaType
	}
}
