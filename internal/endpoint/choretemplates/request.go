package choretemplates

import (
	"errors"

	"github.com/labstack/echo/v5"
)

var errInvalidChoreTemplateID = errors.New("invalid chore template id")

func parseChoreTemplateID(ctx *echo.Context) (int, error) {
	var id int
	err := echo.PathValuesBinder(ctx).Int("id", &id).BindError()
	if err != nil || id <= 0 {
		return 0, errInvalidChoreTemplateID
	}
	return id, nil
}
