package shiftbellapi

type Link struct {
	Href string `json:"href"`
}

type Links map[string]Link

type ActionField struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Required  bool   `json:"required"`
	MaxLength int    `json:"max_length"`
}

type Action struct {
	Href        string        `json:"href"`
	Method      string        `json:"method"`
	ContentType string        `json:"content_type"`
	Fields      []ActionField `json:"fields"`
}

type Actions map[string]Action

type Home struct {
	Links Links `json:"_links"`
}

type GetHomeResult struct {
	Home Home
}

type ChoreTemplate struct {
	Id            int     `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	DeactivatedAt *string `json:"deactivated_at"`
	Links         Links   `json:"_links"`
}

type ChoreTemplateCollection struct {
	Items   []ChoreTemplate `json:"items"`
	More    bool            `json:"more"`
	Links   Links           `json:"_links"`
	Actions Actions         `json:"_actions"`
}

type BrowseChoreTemplatesParams struct {
	Href   string
	Search string
	State  string
	Limit  int
}

type GetChoreTemplatesResult struct {
	Collection ChoreTemplateCollection
}

type CreateChoreTemplateParams struct {
	Action      Action `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateChoreTemplateResponse struct {
	ChoreTemplate ChoreTemplate
	Actions       Actions
	Location      string
}

type CreateChoreTemplateResult struct {
	StatusCode      int
	SuccessResponse *CreateChoreTemplateResponse
	ErrorResponse   *ErrorResponse
}

type GetChoreTemplateParams struct {
	Link Link
}

type GetChoreTemplateResponse struct {
	ChoreTemplate ChoreTemplate
	Actions       Actions
}

type GetChoreTemplateResult struct {
	StatusCode      int
	SuccessResponse *GetChoreTemplateResponse
	ErrorResponse   *ErrorResponse
}

type EditChoreTemplateParams struct {
	Action      Action `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type EditChoreTemplateResponse struct {
	ChoreTemplate ChoreTemplate
	Actions       Actions
}

type EditChoreTemplateResult struct {
	StatusCode      int
	SuccessResponse *EditChoreTemplateResponse
	ErrorResponse   *ErrorResponse
}

type DeactivateChoreTemplateParams struct {
	Action Action
}

type DeactivateChoreTemplateResult struct {
	ChoreTemplate ChoreTemplate
	Actions       Actions
}

type ErrorResponse struct {
	Error   string  `json:"error"`
	Links   Links   `json:"_links"`
	Actions Actions `json:"_actions"`
}

type choreTemplateRepresentation struct {
	ChoreTemplate
	Actions Actions `json:"_actions"`
}
