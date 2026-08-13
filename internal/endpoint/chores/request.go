package chores

import (
	"errors"

	"github.com/labstack/echo/v5"
)

var (
	errInvalidChoreID         = errors.New("invalid chore id")
	errInvalidChoreTemplateID = errors.New("invalid chore template id")
)

func parseChoreID(ctx *echo.Context) (int, error) {
	var id int
	err := echo.PathValuesBinder(ctx).Int("id", &id).BindError()
	if err != nil || id <= 0 {
		return 0, errInvalidChoreID
	}
	return id, nil
}
