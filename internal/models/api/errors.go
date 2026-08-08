package api

type ErrorResponse struct {
	Error   string            `json:"error"`
	Links   map[string]Link   `json:"_links"`
	Actions map[string]Action `json:"_actions"`
}
