package demodata

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	models "github.com/bpsoos/shiftbell/internal/models/demo"
)

func Load(path string) (*models.Data, error) {
	if path == "" {
		return nil, errors.New("DEMO_DATA_FILEPATH is required")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read demo data %q: %w", path, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var data *models.Data
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("decode demo data %q: %w", path, err)
	}
	if data == nil {
		return nil, errors.New("demo data must be a JSON object")
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, errors.New("demo data must contain a single JSON object")
	}
	for i, chore := range data.OneOffChores {
		if strings.TrimSpace(chore.Name) == "" {
			return nil, fmt.Errorf("one_off_chores[%d].name is required", i)
		}
		if chore.CompletedDays != nil && *chore.CompletedDays > 0 {
			return nil, fmt.Errorf(
				"one_off_chores[%d].completed_days cannot be in the future",
				i,
			)
		}
	}
	for i, chore := range data.ScheduledChores {
		if strings.TrimSpace(chore.Name) == "" {
			return nil, fmt.Errorf("scheduled_chores[%d].name is required", i)
		}
		if strings.TrimSpace(chore.ScheduleName) == "" {
			return nil, fmt.Errorf("scheduled_chores[%d].schedule_name is required", i)
		}
		if chore.IntervalDays < 1 {
			return nil, fmt.Errorf(
				"scheduled_chores[%d].interval_days must be positive",
				i,
			)
		}
	}
	return data, nil
}
