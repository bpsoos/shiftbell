package choretemplates

import choretemplateapimodels "github.com/bpsoos/shiftbell/internal/models/api/choretemplates"

type Detail struct {
	ChoreTemplate  choretemplateapimodels.Representation
	CollectionHref string
	EditHref       string
	Notice         string
}
