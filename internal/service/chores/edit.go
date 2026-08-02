package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) Edit(
	ctx context.Context,
	input *models.EditChoreParams,
) (*models.ChoreDetails, error) {
	name, err := s.normalizer.NormalizeName(input.Name)
	if err != nil {
		return nil, validationerrors.ErrInvalidName
	}
	description, err := s.normalizer.NormalizeDescription(input.Description)
	if err != nil {
		return nil, validationerrors.ErrInvalidDescription
	}
	if input.ScheduleId != nil {
		result, err := s.persister.EditScheduled(ctx, &models.EditScheduledChoreParams{
			Id:                      input.Id,
			Name:                    name,
			Description:             description,
			AlsoUpdateChoreTemplate: input.AlsoUpdateChoreTemplate,
		})
		if err != nil {
			return nil, fmt.Errorf("edit scheduled chore: %w", err)
		}

		return result, nil
	}
	if input.Deadline.IsZero() {
		return nil, validationerrors.ErrInvalidDeadline
	}

	result, err := s.persister.EditOneOff(ctx, &models.EditOneOffChoreParams{
		Id:          input.Id,
		Name:        name,
		Description: description,
		Deadline:    input.Deadline,
	})
	if err != nil {
		return nil, fmt.Errorf("edit one-off chore: %w", err)
	}

	return result, nil
}
