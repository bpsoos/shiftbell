package schedules

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
)

func (s *Service) Deactivate(ctx context.Context, id int) (*models.Schedule, error) {
	schedule, err := s.persister.Deactivate(ctx, &models.DeactivateScheduleParams{
		Id:            id,
		DeactivatedAt: s.now(),
	})
	if err != nil {
		return nil, fmt.Errorf("deactivate schedule: %w", err)
	}

	return schedule, nil
}
