package chores

import choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"

type Detail struct {
	HasMultipleProfiles bool
	Chore               choreapimodels.Representation
	BackHref            string
	EditHref            string
	DeleteHref          string
	Notice              string
}
