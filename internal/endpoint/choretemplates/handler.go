package choretemplates

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choretemplateapimodels "github.com/bpsoos/shiftbell/internal/models/api/choretemplates"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choretemplateviewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
)

type Service interface {
	Browse(
		context.Context,
		*models.BrowseChoreTemplatesParams,
	) (*models.ChoreTemplatePage, error)
	Create(
		context.Context,
		*models.CreateChoreTemplateParams,
	) (*models.ChoreTemplate, error)
	Get(context.Context, int) (*models.ChoreTemplateDetails, error)
	Edit(
		context.Context,
		*models.EditChoreTemplateParams,
	) (*models.ChoreTemplate, error)
	Deactivate(context.Context, int) (*models.ChoreTemplate, error)
}

type HandlerDeps struct {
	Service Service
	View    View
}

type Handler struct {
	service Service
	view    View
}

type View interface {
	Collection(choretemplateviewmodels.Collection, bool) templ.Component
	Picker(choretemplateviewmodels.Picker, bool) templ.Component
	Detail(choretemplateviewmodels.Detail, bool) templ.Component
	Error(choretemplateviewmodels.Error, bool) templ.Component
}

type (
	response                 = choretemplateapimodels.Response
	representation           = choretemplateapimodels.Representation
	collectionResponse       = choretemplateapimodels.CollectionResponse
	pickerItemResponse       = choretemplateapimodels.PickerItemResponse
	pickerCollectionResponse = choretemplateapimodels.PickerCollectionResponse
	errorResponse            = api.ErrorResponse
)

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{service: deps.Service, view: deps.View}
}

func newResponse(choreTemplate *models.ChoreTemplate) response {
	return response{
		Id:            choreTemplate.Id,
		Name:          choreTemplate.Name,
		Description:   choreTemplate.Description,
		DeactivatedAt: choreTemplate.DeactivatedAt,
		Links: map[string]api.Link{
			"self":       {Href: fmt.Sprintf("/chore-templates/%d", choreTemplate.Id)},
			"collection": {Href: "/chore-templates"},
		},
	}
}

func newRepresentation(choreTemplate *models.ChoreTemplate) representation {
	response := newResponse(choreTemplate)
	actions := map[string]api.Action{}
	if choreTemplate.DeactivatedAt == nil {
		actions = activeActions(response.Links["self"].Href)
	}
	return representation{Response: response, Actions: actions}
}

func newPickerItemResponse(choreTemplate *models.ChoreTemplate) pickerItemResponse {
	return pickerItemResponse{
		Id:   choreTemplate.Id,
		Name: choreTemplate.Name,
		Select: api.Link{
			Href: fmt.Sprintf("/chores/new?template_id=%d", choreTemplate.Id),
		},
	}
}

func activeActions(selfHref string) map[string]api.Action {
	return map[string]api.Action{
		"edit": {
			Href:        selfHref,
			Method:      http.MethodPatch,
			ContentType: "application/json",
			Fields: []api.ActionField{
				{Name: "name", Type: "string", Required: true},
				{Name: "description", Type: "string", Required: false},
			},
		},
		"deactivate": {
			Href:   selfHref + "/deactivation",
			Method: http.MethodPut,
		},
	}
}

func collectionLink() map[string]api.Link {
	return map[string]api.Link{
		"collection": {Href: "/chore-templates"},
	}
}
