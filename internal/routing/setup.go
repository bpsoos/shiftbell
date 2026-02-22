package routing

import "github.com/labstack/echo/v5"

func (r *Router) Setup(app *echo.Echo) error {
	app.GET("/", r.choreTypeHandler.GetBatch)
	app.GET("/chores", r.choreTypeHandler.GetBatch)
	app.POST("/chores", r.choreTypeHandler.Create)
	return nil
}
