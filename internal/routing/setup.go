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
	app.GET("/schedules", r.scheduleHandler.Browse)
	app.GET("/schedules/:id", r.scheduleHandler.Get)
	app.GET("/schedules/:id/edit", r.scheduleHandler.EditForm)
	app.GET("/schedules/:id/deactivation", r.scheduleHandler.ConfirmDeactivation)
	app.PATCH("/schedules/:id", r.scheduleHandler.Edit)
	app.PUT("/schedules/:id/deactivation", r.scheduleHandler.Deactivate)
	app.GET("/chores", r.choreHandler.GetBatch)
	app.GET("/chores/new", r.choreHandler.New)
	app.POST("/chores", r.choreHandler.Create)
	app.GET("/chores/:id", r.choreHandler.Get)
	app.GET("/chores/:id/edit", r.choreHandler.Edit)
	app.PATCH("/chores/:id", r.choreHandler.Patch)
	app.DELETE("/chores/:id", r.choreHandler.Delete)
	app.GET("/chores/:id/completion", r.choreHandler.ConfirmCompletion)
	app.PUT("/chores/:id/completion", r.choreHandler.Complete)
	app.PATCH("/chores/:id/completion", r.choreHandler.CorrectCompletion)
	app.GET("/chores/:id/completion/edit", r.choreHandler.CompletionCorrectionForm)
	app.GET("/chores/:id/deletion", r.choreHandler.ConfirmDeletion)

	app.POST("/chore-templates", r.choreTemplateHandler.Create)
	app.GET("/chore-templates", r.choreTemplateHandler.Browse)
	app.GET("/chore-templates/new", r.choreTemplateHandler.CreateForm)
	app.GET("/chore-templates/:id", r.choreTemplateHandler.Get)
	app.GET("/chore-templates/:id/schedules", r.choreTemplateHandler.ListSchedules)
	app.GET("/chore-templates/:id/edit", r.choreTemplateHandler.EditForm)
	app.PATCH("/chore-templates/:id", r.choreTemplateHandler.Edit)
	app.GET(
		"/chore-templates/:id/deactivation",
		r.choreTemplateHandler.ConfirmDeactivation,
	)
	app.PUT("/chore-templates/:id/deactivation", r.choreTemplateHandler.Deactivate)
	return nil
}
