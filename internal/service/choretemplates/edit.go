package choretemplates

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	serviceerrors "github.com/bpsoos/shiftbell/internal/service"
)

func (s *Service) Edit(ctx context.Context, params *models.EditChoreTemplateParams) (*models.ChoreTemplate, error) {
	if params == nil {
		return nil, serviceerrors.ErrInvalidName
	}

	name, valid := s.normalizer.NormalizeName(params.Name)
	if !valid {
		return nil, serviceerrors.ErrInvalidName
	}
	description, valid := s.normalizer.NormalizeDescription(params.Description)
	if !valid {
		return nil, serviceerrors.ErrInvalidDescription
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
