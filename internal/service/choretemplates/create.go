package choretemplates

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) Create(ctx context.Context, params *models.CreateChoreTemplateParams) (*models.ChoreTemplate, error) {
	if params == nil {
		return nil, validationerrors.ErrInvalidName
	}

	name, err := s.normalizer.NormalizeName(params.Name)
	if err != nil {
		return nil, validationerrors.ErrInvalidName
	}
	description, err := s.normalizer.NormalizeDescription(params.Description)
	if err != nil {
		return nil, validationerrors.ErrInvalidDescription
	}

	choreTemplate, err := s.persister.Create(ctx, &models.CreateChoreTemplateParams{
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("create chore template: %w", err)
	}

	return choreTemplate, nil
}
