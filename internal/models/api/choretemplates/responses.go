package choretemplates

import (
	"time"

	api "github.com/bpsoos/shiftbell/internal/models/api"
)

type Response struct {
	Id            int           `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	DeactivatedAt *time.Time    `json:"deactivated_at"`
	Links         api.Relations `json:"_links"`
}

type Representation struct {
	Response
	Actions api.Relations `json:"_actions"`
}

type CollectionResponse struct {
	Items   []Response    `json:"items"`
	More    bool          `json:"more"`
	Links   api.Relations `json:"_links"`
	Actions api.Relations `json:"_actions"`
}

type PickerItemResponse struct {
	Id    int           `json:"id"`
	Name  string        `json:"name"`
	Links api.Relations `json:"_links"`
}

type PickerCollectionResponse struct {
	Items []PickerItemResponse `json:"items"`
	More  bool                 `json:"more"`
	Links api.Relations        `json:"_links"`
}
