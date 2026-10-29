package populate

import (
	"context"
	"fmt"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	demomodels "github.com/bpsoos/shiftbell/internal/models/demo"
)

func (s *Service) Populate(ctx context.Context, input *demomodels.Data) error {
	hasData, err := s.dataPersister.HasData(ctx)
	if err != nil {
		return fmt.Errorf("check for existing demo data: %w", err)
	}
	if hasData {
		return nil
	}
	now := s.now().In(s.timezone)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	for _, chore := range input.OneOffChores {
		created, err := s.chorePersister.CreateManualOneOff(
			ctx,
			&models.CreateManualOneOffParams{
				Name:                chore.Name,
				Description:         chore.Description,
				Deadline:            today.AddDate(0, 0, chore.DeadlineDays),
				SaveAsChoreTemplate: chore.SaveAsChoreTemplate,
			},
		)
		if err != nil {
			return fmt.Errorf("create demo chore %q: %w", chore.Name, err)
		}
		if chore.CompletedDays != nil {
			if err := s.chorePersister.MarkComplete(
				created.Chore.Id,
				today.AddDate(0, 0, *chore.CompletedDays),
			); err != nil {
				return fmt.Errorf("complete demo chore %q: %w", chore.Name, err)
			}
		}
	}
	for _, chore := range input.ScheduledChores {
		if _, err := s.chorePersister.CreateManualScheduled(
			ctx,
			&models.CreateManualScheduledParams{
				Name:         chore.Name,
				Description:  chore.Description,
				Deadline:     today.AddDate(0, 0, chore.DeadlineDays),
				ScheduleName: chore.ScheduleName,
				IntervalDays: chore.IntervalDays,
			},
		); err != nil {
			return fmt.Errorf("create scheduled demo chore %q: %w", chore.Name, err)
		}
	}
	return nil
}
