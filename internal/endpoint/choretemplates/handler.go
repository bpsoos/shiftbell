package choretemplates

import (
	"context"
	"fmt"
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

type collectionResponse struct {
	Items   []response                   `json:"items"`
	More    bool                         `json:"more"`
	Links   map[string]hypermedia.Link   `json:"_links"`
	Actions map[string]hypermedia.Action `json:"_actions"`
}

type errorResponse struct {
	Error string `json:"error"`
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
