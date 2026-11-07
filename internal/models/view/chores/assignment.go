package chores

import (
	choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
	profilemodels "github.com/bpsoos/shiftbell/internal/models/profiles"
)

type Assignment struct {
	Chore     choreapimodels.Representation
	Profiles  []profilemodels.Profile
	Error     string
	RetryHref string
}
