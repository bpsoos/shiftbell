package profiles

import models "github.com/bpsoos/shiftbell/internal/models/profiles"

type Selection struct {
	Profiles          []models.Profile
	SelectedProfileId int
	ReturnTo          string
	Error             string
}

type Navigation struct {
	Profile  models.Profile
	Profiles []models.Profile
}
