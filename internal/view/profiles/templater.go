package profiles

import (
	"time"

	"github.com/a-h/templ"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/bpsoos/shiftbell/internal/view/layouts"
)

type Config struct {
	AppTimezone       *time.Location
	ProfilesUIEnabled bool
}

type Templater struct {
	appTimezone       *time.Location
	profilesUIEnabled bool
}

func NewTemplater(config Config) *Templater {
	return &Templater{
		appTimezone:       config.AppTimezone,
		profilesUIEnabled: config.ProfilesUIEnabled,
	}
}

func (t *Templater) Collection(
	model viewmodels.Collection,
	fullPage bool,
) templ.Component {
	return layouts.Frame(
		"profiles",
		fullPage,
		t.appTimezone.String(),
		t.profilesUIEnabled,
		collection(model),
	)
}

func (t *Templater) Selection(model viewmodels.Selection, fullPage bool) templ.Component {
	return layouts.Frame(
		"profile-selection",
		fullPage,
		t.appTimezone.String(),
		t.profilesUIEnabled,
		selection(model),
	)
}

func (t *Templater) SelectedProfile(model viewmodels.Navigation) templ.Component {
	return selectedProfile(model)
}

func (t *Templater) Form(model viewmodels.Form, fullPage bool) templ.Component {
	if model.Setup {
		return layouts.Frame(
			"setup",
			fullPage,
			t.appTimezone.String(),
			t.profilesUIEnabled,
			setupForm(model),
		)
	}
	return layouts.Frame(
		"profiles",
		fullPage,
		t.appTimezone.String(),
		t.profilesUIEnabled,
		profileForm(model),
	)
}

func (t *Templater) Error(model viewmodels.Error, fullPage bool) templ.Component {
	return layouts.Frame(
		"profiles",
		fullPage,
		t.appTimezone.String(),
		t.profilesUIEnabled,
		errorContent(model),
	)
}
