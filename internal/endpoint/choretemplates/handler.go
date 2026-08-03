package choretemplates

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
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
}

type Handler struct {
	service Service
}

type response struct {
	Id            int                        `json:"id"`
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	DeactivatedAt *time.Time                 `json:"deactivated_at"`
	Links         map[string]hypermedia.Link `json:"_links"`
}

type representation struct {
	response
	Actions map[string]hypermedia.Action `json:"_actions"`
}

type collectionResponse struct {
	Items   []response                   `json:"items"`
	More    bool                         `json:"more"`
	Links   map[string]hypermedia.Link   `json:"_links"`
	Actions map[string]hypermedia.Action `json:"_actions"`
}

type errorResponse struct {
	Error   string                       `json:"error"`
	Links   map[string]hypermedia.Link   `json:"_links"`
	Actions map[string]hypermedia.Action `json:"_actions"`
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{service: deps.Service}
}

func newResponse(choreTemplate *models.ChoreTemplate) response {
	return response{
		Id:            choreTemplate.Id,
		Name:          choreTemplate.Name,
		Description:   choreTemplate.Description,
		DeactivatedAt: choreTemplate.DeactivatedAt,
		Links: map[string]hypermedia.Link{
			"self":       {Href: fmt.Sprintf("/chore-templates/%d", choreTemplate.Id)},
			"collection": {Href: "/chore-templates"},
		},
	}
}

func newRepresentation(choreTemplate *models.ChoreTemplate) representation {
	response := newResponse(choreTemplate)
	actions := map[string]hypermedia.Action{}
	if choreTemplate.DeactivatedAt == nil {
		actions = activeActions(response.Links["self"].Href)
	}
	return representation{response: response, Actions: actions}
}

func activeActions(selfHref string) map[string]hypermedia.Action {
	return map[string]hypermedia.Action{
		"edit": {
			Href:        selfHref,
			Method:      http.MethodPatch,
			ContentType: "application/json",
			Fields: []hypermedia.ActionField{
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

func collectionLink() map[string]hypermedia.Link {
	return map[string]hypermedia.Link{
		"collection": {Href: "/chore-templates"},
	}
}
