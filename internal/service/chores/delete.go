package chores

import (
	"context"
	"fmt"
)

func (s *Service) Delete(ctx context.Context, id int) error {
	if err := s.persister.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete chore: %w", err)
	}

	return nil
}
