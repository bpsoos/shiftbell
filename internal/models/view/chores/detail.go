package chores

import choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"

type Detail struct {
	Chore    choreapimodels.Representation
	BackHref string
	EditHref string
	Notice   string
}
