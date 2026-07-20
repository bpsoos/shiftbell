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
	app.GET("/chores/new", r.choreHandler.New)
	app.POST("/chores", r.choreHandler.Create)
	app.GET("/chores/:id", r.choreHandler.Get)
	app.PATCH("/chores/:id", r.choreHandler.Patch)

	app.GET("/choretemplates", r.choreTemplateHandler.GetBatch)
	app.POST("/choretemplates", r.choreTemplateHandler.Create)
	app.DELETE("/choretemplates/:id", r.choreTemplateHandler.Delete)
	return nil
}
