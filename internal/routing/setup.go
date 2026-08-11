package routing

import (
	"fmt"

	"github.com/bpsoos/shiftbell/internal/static"
	"github.com/labstack/echo/v5"
)

func (r *Router) Setup(app *echo.Echo) error {
	assetsFS, err := static.Assets()
	if err != nil {
		return fmt.Errorf("assets: %w", err)
	}
	app.StaticFS("/assets", assetsFS)

	app.GET("/", r.homeHandler.Get)
	app.GET("/chores", r.choreHandler.GetBatch)
	app.GET("/chores/new", r.choreHandler.New)
	app.POST("/chores", r.choreHandler.Create)
	app.GET("/chores/:id", r.choreHandler.Get)
	app.PATCH("/chores/:id", r.choreHandler.Patch)
	app.DELETE("/chores/:id", r.choreHandler.Delete)
	app.GET("/chores/:id/completion", r.choreHandler.ConfirmCompletion)
	app.PUT("/chores/:id/completion", r.choreHandler.Complete)
	app.PATCH("/chores/:id/completion", r.choreHandler.CorrectCompletion)

	app.POST("/chore-templates", r.choreTemplateHandler.Create)
	app.GET("/chore-templates", r.choreTemplateHandler.Browse)
	app.GET("/chore-templates/:id", r.choreTemplateHandler.Get)
	app.PATCH("/chore-templates/:id", r.choreTemplateHandler.Edit)
	app.PUT("/chore-templates/:id/deactivation", r.choreTemplateHandler.Deactivate)
	return nil
}
