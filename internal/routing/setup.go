package routing

import (
	"github.com/labstack/echo/v5"
)

func (r *Router) Setup(app *echo.Echo) error {
	app.GET("/", r.choreHandler.GetBatch)
	app.GET("/chores", r.choreHandler.GetBatch)
	app.PATCH("/chores/:id/status", r.choreHandler.PatchStatus)
	app.GET("/choretypes", r.choreTypeHandler.GetBatch)
	app.POST("/choretypes", r.choreTypeHandler.Create)
	return nil
}
