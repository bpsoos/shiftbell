package chores

import (
	"context"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/models"
	serviceerrors "github.com/bpsoos/shiftbell/internal/service"
)

func (s *Service) Create(ctx context.Context, input *models.CreateChoreInput) (*models.CreateChoreResult, error) {
	if input.Deadline.IsZero() {
		return nil, serviceerrors.ErrInvalidDeadline
	}

	if s.hasInvalidInterval(input) {
		return nil, serviceerrors.ErrInvalidInterval
	}

	if s.isTemplateScheduled(input) {
		scheduleName, valid := s.normalizer.NormalizeName(input.ScheduleName)
		if !valid {
			return nil, serviceerrors.ErrInvalidName
		}

		result, err := s.persister.CreateTemplateScheduled(ctx, &models.CreateTemplateScheduledInput{
			ChoreTemplateId: *input.ChoreTemplateId,
			Deadline:        input.Deadline,
			ScheduleName:    scheduleName,
			IntervalDays:    *input.IntervalDays,
		})
		if err != nil {
			return nil, fmt.Errorf("create template scheduled chore: %w", err)
		}

		return result, nil
	}

	name, valid := s.normalizer.NormalizeName(input.Name)
	if !valid {
		return nil, serviceerrors.ErrInvalidName
	}
	description, valid := s.normalizer.NormalizeDescription(input.Description)
	if !valid {
		return nil, serviceerrors.ErrInvalidDescription
	}
	if s.isScheduled(input) {
		scheduleName, valid := s.normalizer.NormalizeName(input.ScheduleName)
		if !valid {
			return nil, serviceerrors.ErrInvalidName
		}

		result, err := s.persister.CreateManualScheduled(ctx, &models.CreateManualScheduledInput{
			Name:         name,
			Description:  description,
			Deadline:     input.Deadline,
			ScheduleName: scheduleName,
			IntervalDays: *input.IntervalDays,
		})
		if err != nil {
			return nil, fmt.Errorf("create manual scheduled chore: %w", err)
		}

		return result, nil
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

func (s *Service) hasInvalidInterval(input *models.CreateChoreInput) bool {
	return input.IntervalDays != nil && (*input.IntervalDays < 1 || *input.IntervalDays > 3650)
}

func (s *Service) isTemplateScheduled(input *models.CreateChoreInput) bool {
	return input.ChoreTemplateId != nil && s.isScheduled(input)
}

func (s *Service) isScheduled(input *models.CreateChoreInput) bool {
	return input.IntervalDays != nil
}
