package choretemplates

import (
	"strings"
	"time"

	"github.com/a-h/templ"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	collectioncontrols "github.com/bpsoos/shiftbell/internal/view/collectioncontrols"
	confirmationview "github.com/bpsoos/shiftbell/internal/view/confirmation"
	"github.com/bpsoos/shiftbell/internal/view/layouts"
)

type Config struct {
	AppTimezone *time.Location
}

type Templater struct {
	appTimezone *time.Location
}

func NewTemplater(config Config) *Templater {
	return &Templater{appTimezone: config.AppTimezone}
}

func (t *Templater) Collection(
	model viewmodels.Collection,
	fullPage bool,
) templ.Component {
	if fullPage && model.SearchOpen {
		model.AutofocusSearch = true
	}
	return layouts.Frame("chore-templates", fullPage, collection(model))
}

func templateCollectionViewHref(filter models.ChoreTemplateFilter) string {
	if filter == models.ChoreTemplateFilterDeactivated {
		return "/chore-templates?state=deactivated"
	}
	return "/chore-templates"
}

func templateSearchHref(filter models.ChoreTemplateFilter) string {
	if filter == models.ChoreTemplateFilterDeactivated {
		return "/chore-templates?state=deactivated&search="
	}
	return "/chore-templates?search="
}

func templateSearchPlaceholder(filter models.ChoreTemplateFilter) string {
	if filter == models.ChoreTemplateFilterDeactivated {
		return "Search deactivated templates"
	}
	return "Search templates"
}

func pickerControlModel(model viewmodels.Picker) collectioncontrols.Model {
	return collectioncontrols.Model{
		ActionHref:      "/chore-templates",
		AutofocusSearch: model.AutofocusSearch,
		HiddenInputs: []collectioncontrols.HiddenInput{
			{Name: "picker", Value: "1"},
		},
		Search:            model.Search,
		SearchInputID:     "template-picker-search",
		SearchPlaceholder: "Search templates",
	}
}

func (t *Templater) Picker(
	model viewmodels.Picker,
	fullPage bool,
) templ.Component {
	return layouts.Frame("chores", fullPage, picker(model))
}

func (t *Templater) Detail(
	model viewmodels.Detail,
	fullPage bool,
) templ.Component {
	return layouts.Frame("chore-templates", fullPage, detail(t, model))
}

func (t *Templater) EditForm(
	model viewmodels.EditForm,
	fullPage bool,
) templ.Component {
	return layouts.Frame("chore-templates", fullPage, editForm(model))
}

func (t *Templater) ConfirmationDialog(
	model confirmationviewmodels.Dialog,
) templ.Component {
	return confirmationview.Dialog(model)
}

func (t *Templater) Error(
	model viewmodels.Error,
	fullPage bool,
) templ.Component {
	return layouts.Frame("chore-templates", fullPage, errorContent(model))
}

func (t *Templater) formatTime(value time.Time) string {
	return value.In(t.appTimezone).Format("2 Jan 2006, 15:04 MST")
}

func knownLink(links api.Relations, relation string) string {
	return links.Href(relation)
}

func oneOffCreationHref(href string) string {
	if strings.Contains(href, "?") {
		return href + "&recurrence=one-off"
	}
	return href + "?recurrence=one-off"
}

func formControlClass(className string, errorMessage string) string {
	if errorMessage != "" {
		className += " is-invalid"
	}
	return className
}
