package choretemplates

import (
	"time"

	api "github.com/bpsoos/shiftbell/internal/models/api"
)

type Response struct {
	Id            int                 `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	DeactivatedAt *time.Time          `json:"deactivated_at"`
	Links         map[string]api.Link `json:"_links"`
}

type Representation struct {
	Response
	Actions map[string]api.Action `json:"_actions"`
}

type CollectionResponse struct {
	Items   []Response            `json:"items"`
	More    bool                  `json:"more"`
	Links   map[string]api.Link   `json:"_links"`
	Actions map[string]api.Action `json:"_actions"`
}

type PickerItemResponse struct {
	Id     int      `json:"id"`
	Name   string   `json:"name"`
	Select api.Link `json:"select"`
}

type PickerCollectionResponse struct {
	Items []PickerItemResponse `json:"items"`
	More  bool                 `json:"more"`
	Links map[string]api.Link  `json:"_links"`
}
