package chores

import (
	"github.com/a-h/templ"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
)

func (t *Templater) Assignment(viewmodels.Assignment) templ.Component {
	return templ.Raw(`<span data-assignment-placeholder></span>`)
}
