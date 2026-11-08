package profiles

import "github.com/bpsoos/shiftbell/internal/models/api"

type Response struct {
	Id      int           `json:"id"`
	Name    string        `json:"name"`
	Links   api.Relations `json:"_links"`
	Actions api.Relations `json:"_actions"`
}

type Collection struct {
	Items   []Response    `json:"items"`
	Links   api.Relations `json:"_links"`
	Actions api.Relations `json:"_actions"`
}

type Setup struct {
	Required bool          `json:"required"`
	Links    api.Relations `json:"_links"`
	Actions  api.Relations `json:"_actions"`
}
