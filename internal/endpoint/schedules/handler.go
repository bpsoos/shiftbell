package schedules

import (
	"context"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
)

type Service interface {
	Browse(context.Context, *models.BrowseSchedulesParams) (*models.SchedulePage, error)
	Get(context.Context, int) (*models.ScheduleDetails, error)
	Edit(context.Context, *models.EditScheduleParams) (*models.Schedule, error)
}

type HandlerDeps struct {
	Service Service
}

type Handler struct {
	service Service
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{service: deps.Service}
}
