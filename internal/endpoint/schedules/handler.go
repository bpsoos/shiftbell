package schedules

import (
	"context"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
)

type Service interface {
	Browse(context.Context, *models.BrowseSchedulesParams) (*models.SchedulePage, error)
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
