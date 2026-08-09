package choretemplates

import (
	"strings"
	"time"

	"github.com/a-h/templ"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
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
	return layouts.Frame("chore-templates", fullPage, collection(model))
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
