package choretemplates

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) Edit(
	ctx context.Context,
	params *models.EditChoreTemplateParams,
) (*models.ChoreTemplate, error) {
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

	choreTemplate, err := s.persister.Edit(ctx, &models.EditChoreTemplateParams{
		Id:          params.Id,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("edit chore template: %w", err)
	}

	return choreTemplate, nil
}
