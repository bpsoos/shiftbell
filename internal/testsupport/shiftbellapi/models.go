package shiftbellapi

type Link struct {
	Href string `json:"href"`
}

type Links map[string]Link

type ActionField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
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

type Chore struct {
	Id          int     `json:"id"`
	ScheduleId  *int    `json:"schedule_id"`
	Status      string  `json:"status"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Deadline    string  `json:"deadline"`
	CompletedOn *string `json:"completed_on"`
	Links       Links   `json:"_links"`
}

type ChoreCollection struct {
	Items   []Chore `json:"items"`
	More    bool    `json:"more"`
	Links   Links   `json:"_links"`
	Actions Actions `json:"_actions"`
}

type GetChoresResult struct {
	Collection ChoreCollection
}

type ChoreCreationChoice struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type ChoreCreationTemplate struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ChoreCreationStep struct {
	Step     string                 `json:"step"`
	Template *ChoreCreationTemplate `json:"template,omitempty"`
	Choices  []ChoreCreationChoice  `json:"choices,omitempty"`
	Fields   []ActionField          `json:"fields,omitempty"`
	Action   *Action                `json:"action,omitempty"`
}

type GetChoreCreationStepResult struct {
	Step ChoreCreationStep
}

type CreateChoreParams struct {
	Action              Action `json:"-"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	Deadline            string `json:"deadline"`
	ChoreTemplateId     *int   `json:"chore_template_id,omitempty"`
	ScheduleName        string `json:"schedule_name,omitempty"`
	IntervalDays        *int   `json:"interval_days,omitempty"`
	SaveAsChoreTemplate bool   `json:"save_as_chore_template"`
}

type CreateChoreResponse struct {
	Chore    Chore
	Actions  Actions
	Location string
}

type CreateChoreResult struct {
	StatusCode      int
	SuccessResponse *CreateChoreResponse
	ErrorResponse   *ErrorResponse
}

type GetChoreParams struct {
	Link Link
}

type GetChoreResponse struct {
	Chore   Chore
	Actions Actions
}

type GetChoreResult struct {
	StatusCode      int
	SuccessResponse *GetChoreResponse
	ErrorResponse   *ErrorResponse
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

type choreRepresentation struct {
	Chore
	Actions Actions `json:"_actions"`
}
