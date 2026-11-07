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

	app.GET("/setup", r.profileHandler.SetupStatus)
	app.POST("/setup", r.profileHandler.Setup)
	app.GET("/profiles/selection/status", r.profileHandler.SelectionStatus)
	app.GET("/profiles/selection", r.profileHandler.Selection)
	app.POST("/profiles/selection", r.profileHandler.Select)
	pages := app.Group("")
	pages.GET("/", r.homeHandler.Get)
	pages.GET("/profiles", r.profileHandler.List)
	pages.POST("/profiles", r.profileHandler.Create)
	pages.GET("/profiles/new", r.profileHandler.CreateForm)
	pages.GET("/schedules", r.scheduleHandler.Browse)
	pages.GET("/schedules/:id", r.scheduleHandler.Get)
	pages.GET("/schedules/:id/edit", r.scheduleHandler.EditForm)
	pages.GET("/schedules/:id/deactivation", r.scheduleHandler.ConfirmDeactivation)
	pages.PATCH("/schedules/:id", r.scheduleHandler.Edit)
	pages.PUT("/schedules/:id/deactivation", r.scheduleHandler.Deactivate)
	pages.GET("/chores", r.choreHandler.GetBatch)
	pages.GET("/chores/new", r.choreHandler.New)
	pages.POST("/chores", r.choreHandler.Create)
	pages.GET("/chores/:id", r.choreHandler.Get)
	pages.PUT("/chores/:id/assignment", r.choreHandler.SetAssignment)
	pages.DELETE("/chores/:id/assignment", r.choreHandler.ClearAssignment)
	pages.GET("/chores/:id/edit", r.choreHandler.Edit)
	pages.PATCH("/chores/:id", r.choreHandler.Patch)
	pages.DELETE("/chores/:id", r.choreHandler.Delete)
	pages.GET("/chores/:id/completion", r.choreHandler.ConfirmCompletion)
	pages.PUT("/chores/:id/completion", r.choreHandler.Complete)
	pages.PATCH("/chores/:id/completion", r.choreHandler.CorrectCompletion)
	pages.GET("/chores/:id/completion/edit", r.choreHandler.CompletionCorrectionForm)
	pages.GET("/chores/:id/deletion", r.choreHandler.ConfirmDeletion)

	pages.POST("/chore-templates", r.choreTemplateHandler.Create)
	pages.GET("/chore-templates", r.choreTemplateHandler.Browse)
	pages.GET("/chore-templates/new", r.choreTemplateHandler.CreateForm)
	pages.GET("/chore-templates/:id", r.choreTemplateHandler.Get)
	pages.GET("/chore-templates/:id/schedules", r.choreTemplateHandler.ListSchedules)
	pages.GET("/chore-templates/:id/edit", r.choreTemplateHandler.EditForm)
	pages.PATCH("/chore-templates/:id", r.choreTemplateHandler.Edit)
	pages.GET(
		"/chore-templates/:id/deactivation",
		r.choreTemplateHandler.ConfirmDeactivation,
	)
	pages.PUT("/chore-templates/:id/deactivation", r.choreTemplateHandler.Deactivate)
	return nil
}
