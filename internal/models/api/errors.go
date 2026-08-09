package api

type ErrorResponse struct {
	Error   string    `json:"error"`
	Links   Relations `json:"_links"`
	Actions Relations `json:"_actions"`
}
