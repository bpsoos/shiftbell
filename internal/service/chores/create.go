package chores

import (
	"context"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/models"
	serviceerrors "github.com/bpsoos/shiftbell/internal/service"
)

func (s *Service) Create(ctx context.Context, input *models.CreateChoreInput) (*models.CreateChoreResult, error) {
	name, valid := s.normalizer.NormalizeName(input.Name)
	if !valid {
		return nil, serviceerrors.ErrInvalidName
	}
	description, valid := s.normalizer.NormalizeDescription(input.Description)
	if !valid {
		return nil, serviceerrors.ErrInvalidDescription
	}

	result, err := s.persister.CreateManualOneOff(ctx, &models.CreateManualOneOffInput{
		Name:                name,
		Description:         description,
		Deadline:            input.Deadline,
		SaveAsChoreTemplate: input.SaveAsChoreTemplate,
	})
	if err != nil {
		return nil, fmt.Errorf("create manual one-off chore: %w", err)
	}

	return result, nil
}
