package choretemplates

import (
	"context"
	"fmt"

	"github.com/a-h/templ"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choretemplateapimodels "github.com/bpsoos/shiftbell/internal/models/api/choretemplates"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choretemplateviewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
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
	EditForm(choretemplateviewmodels.EditForm, bool) templ.Component
	ConfirmationDialog(confirmationviewmodels.Dialog) templ.Component
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
		Links: api.Relations{
			{Rel: "self", Href: fmt.Sprintf("/chore-templates/%d", choreTemplate.Id)},
			{Rel: "collection", Href: "/chore-templates"},
		},
	}
}

func newRepresentation(choreTemplate *models.ChoreTemplate) representation {
	response := newResponse(choreTemplate)
	actions := api.Relations{}
	if choreTemplate.DeactivatedAt == nil {
		actions = activeActions(response.Links.Href("self"))
	}
	return representation{Response: response, Actions: actions}
}

func newPickerItemResponse(choreTemplate *models.ChoreTemplate) pickerItemResponse {
	return pickerItemResponse{
		Id:   choreTemplate.Id,
		Name: choreTemplate.Name,
		Links: api.Relations{
			{
				Rel:  "select",
				Href: fmt.Sprintf("/chores/new?template_id=%d", choreTemplate.Id),
			},
		},
	}
}

func activeActions(selfHref string) api.Relations {
	return api.Relations{
		{Rel: "edit", Href: selfHref},
		{Rel: "deactivate", Href: selfHref + "/deactivation"},
	}
}

func collectionLink() api.Relations {
	return api.Relations{
		{Rel: "collection", Href: "/chore-templates"},
	}
}
