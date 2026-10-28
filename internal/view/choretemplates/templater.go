package choretemplates

import (
	"time"

	"github.com/a-h/templ"
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
	return layouts.Frame(
		"chore-templates",
		fullPage,
		t.appTimezone.String(),
		collection(model),
	)
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
	return layouts.Frame("chores", fullPage, t.appTimezone.String(), picker(model))
}

func (t *Templater) Detail(
	model viewmodels.Detail,
	fullPage bool,
) templ.Component {
	return layouts.Frame(
		"chore-templates",
		fullPage,
		t.appTimezone.String(),
		detail(t, model),
	)
}

func (t *Templater) EditForm(
	model viewmodels.EditForm,
	fullPage bool,
) templ.Component {
	return layouts.Frame(
		"chore-templates",
		fullPage,
		t.appTimezone.String(),
		editForm(model),
	)
}

func (t *Templater) CreateForm(
	model viewmodels.CreateForm,
	fullPage bool,
) templ.Component {
	return layouts.Frame(
		"chore-templates",
		fullPage,
		t.appTimezone.String(),
		createForm(model),
	)
}

func (t *Templater) ConfirmationDialog(
	model confirmationviewmodels.Dialog,
) templ.Component {
	return confirmationview.Dialog(model)
}

func (t *Templater) ScheduleList(
	model viewmodels.ScheduleList,
	fullPage bool,
) templ.Component {
	if model.Inline {
		return scheduleList(model)
	}
	return layouts.Frame(
		"chore-templates",
		fullPage,
		t.appTimezone.String(),
		scheduleListPage(model),
	)
}

func (t *Templater) Error(
	model viewmodels.Error,
	fullPage bool,
) templ.Component {
	return layouts.Frame(
		"chore-templates",
		fullPage,
		t.appTimezone.String(),
		errorContent(model),
	)
}

func (t *Templater) formatTime(value time.Time) string {
	return value.In(t.appTimezone).Format("2006.01.02. 15:04")
}

func formControlClass(className string, errorMessage string) string {
	if errorMessage != "" {
		className += " is-invalid"
	}
	return className
}
