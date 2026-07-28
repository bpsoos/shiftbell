package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) CorrectCompletion(
	ctx context.Context,
	input *models.CorrectCompletionParams,
) (*models.ChoreDetails, error) {
	if !s.validCompletionDate(input.CompletedOn) {
		return nil, validationerrors.ErrInvalidCompletionDate
	}

	result, err := s.persister.CorrectCompletion(ctx, &models.CorrectCompletionParams{
		Id:          input.Id,
		CompletedOn: input.CompletedOn,
	})
	if err != nil {
		return nil, fmt.Errorf("correct chore completion: %w", err)
	}

	return result.Chore, nil
}
