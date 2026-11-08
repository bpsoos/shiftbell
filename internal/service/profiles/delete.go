package profiles

import (
	"context"
	"fmt"
)

func (s *Service) Delete(ctx context.Context, id int) error {
	if err := s.persister.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}
	return nil
}
