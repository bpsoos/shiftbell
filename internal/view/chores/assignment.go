package chores

import (
	"github.com/a-h/templ"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
)

func (t *Templater) Assignment(
	model viewmodels.Assignment,
) templ.Component {
	return assignmentForm(model)
}
