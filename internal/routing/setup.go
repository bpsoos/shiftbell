package routing

import "github.com/labstack/echo/v5"

func (r *Router) Setup(app *echo.Echo) error {
	app.GET("/", r.homeHandler.Home)
	return nil
}
