package chores

import (
	"context"
	"fmt"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) Complete(
	ctx context.Context,
	input *models.CompleteChoreParams,
) (*models.CompleteChoreResult, error) {
	if !s.validCompletionDate(input.CompletedOn) {
		return nil, validationerrors.ErrInvalidCompletionDate
	}

	result, err := s.persister.Complete(ctx, &models.CompleteChoreParams{
		Id:          input.Id,
		CompletedOn: input.CompletedOn,
	})
	if err != nil {
		return nil, fmt.Errorf("complete chore: %w", err)
	}

	return result, nil
}

func (s *Service) validCompletionDate(value time.Time) bool {
	if value.IsZero() {
		return false
	}

	year, month, day := value.Date()
	completedOn := time.Date(year, month, day, 0, 0, 0, 0, s.timezone)
	year, month, day = s.now().In(s.timezone).Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, s.timezone)

	return !completedOn.After(today)
}
