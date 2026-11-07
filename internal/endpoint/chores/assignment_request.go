package chores

import (
	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	models "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
)

type assignmentRequest struct {
	OwnerProfileId int `json:"owner_profile_id" form:"owner_profile_id"`
}

func parseAssignmentRequest(ctx *echo.Context) (int, error) {
	var request assignmentRequest
	if err := binding.Bind(ctx, &request); err != nil {
		return 0, err
	}
	if request.OwnerProfileId <= 0 {
		return 0, models.ErrInvalidOwnerProfile
	}
	return request.OwnerProfileId, nil
}
