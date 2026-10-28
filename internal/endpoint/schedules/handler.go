package schedules

import (
	"context"

	"github.com/a-h/templ"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	confirmationmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
)

type Service interface {
	Browse(context.Context, *models.BrowseSchedulesParams) (*models.SchedulePage, error)
	Get(context.Context, int) (*models.ScheduleDetails, error)
	Edit(context.Context, *models.EditScheduleParams) (*models.Schedule, error)
	Deactivate(context.Context, int) (*models.Schedule, error)
}

type HandlerDeps struct {
	Service Service
	View    View
}

type Handler struct {
	service Service
	view    View
}

type View interface {
	Collection(viewmodels.Collection, bool) templ.Component
	Detail(viewmodels.Detail, bool) templ.Component
	Error(viewmodels.Error, bool) templ.Component
	EditForm(viewmodels.EditForm, bool) templ.Component
	ConfirmationDialog(confirmationmodels.Dialog) templ.Component
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{service: deps.Service, view: deps.View}
}
