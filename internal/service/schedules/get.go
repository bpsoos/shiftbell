package schedules

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
)

func (s *Service) Get(ctx context.Context, id int) (*models.ScheduleDetails, error) {
	details, err := s.persister.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}

	return details, nil
}
