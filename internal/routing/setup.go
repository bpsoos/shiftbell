package routing

import (
	"fmt"

	"github.com/bpsoos/shiftbell/internal/static"
	"github.com/labstack/echo/v5"
)

func (r *Router) Setup(app *echo.Echo) error {
	stylesFs, err := static.Styles()
	if err != nil {
		return fmt.Errorf("styles: %v", err)
	}
	app.StaticFS("/styles", stylesFs)

	app.GET("/", r.choreHandler.GetBatch)
	app.GET("/chores", r.choreHandler.GetBatch)
	app.GET("/chores/:id", r.choreHandler.Get)
	app.PATCH("/chores/:id", r.choreHandler.PatchStatus)

	app.GET("/choretypes", r.choreTypeHandler.GetBatch)
	app.POST("/choretypes", r.choreTypeHandler.Create)
	app.DELETE("/choretypes/:id", r.choreTypeHandler.Delete)
	return nil
}
