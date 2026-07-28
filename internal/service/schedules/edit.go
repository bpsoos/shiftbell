package schedules

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) Edit(ctx context.Context, params *models.EditScheduleParams) (*models.Schedule, error) {
	name, err := s.normalizer.NormalizeName(params.Name)
	if err != nil {
		return nil, validationerrors.ErrInvalidName
	}
	if params.IntervalDays < 1 || params.IntervalDays > 3650 {
		return nil, validationerrors.ErrInvalidInterval
	}

	schedule, err := s.persister.Edit(ctx, &models.EditScheduleParams{
		Id:           params.Id,
		Name:         name,
		IntervalDays: params.IntervalDays,
	})
	if err != nil {
		return nil, fmt.Errorf("edit schedule: %w", err)
	}

	return schedule, nil
}
