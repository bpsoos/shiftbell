package shiftbellapi

type Relation struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type Relations []Relation

func (relations Relations) Href(rel string) string {
	for _, relation := range relations {
		if relation.Rel == rel {
			return relation.Href
		}
	}
	return ""
}

type Home struct {
	Links Relations `json:"_links"`
}

type GetHomeResult struct {
	Home Home
}

type Chore struct {
	Id          int       `json:"id"`
	ScheduleId  *int      `json:"schedule_id"`
	Status      string    `json:"status"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Deadline    string    `json:"deadline"`
	CompletedOn *string   `json:"completed_on"`
	Links       Relations `json:"_links"`
}

type ChoreCollection struct {
	Items   []Chore   `json:"items"`
	More    bool      `json:"more"`
	Links   Relations `json:"_links"`
	Actions Relations `json:"_actions"`
}

type GetChoresResult struct {
	Collection ChoreCollection
}

type BrowseChoresParams struct {
	Href   string
	Search string
	Status string
	Limit  int
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
	Actions  Relations              `json:"_actions"`
}

type GetChoreCreationStepResult struct {
	StatusCode      int
	SuccessResponse *ChoreCreationStep
	ErrorResponse   *ErrorResponse
}

type CreateChoreParams struct {
	Name                string `json:"name,omitempty"`
	Description         string `json:"description,omitempty"`
	Deadline            string `json:"deadline"`
	ChoreTemplateId     *int   `json:"chore_template_id,omitempty"`
	ScheduleName        string `json:"schedule_name,omitempty"`
	IntervalDays        *int   `json:"interval_days,omitempty"`
	SaveAsChoreTemplate bool   `json:"save_as_chore_template,omitempty"`
}

type CreateChoreResponse struct {
	Chore    Chore
	Actions  Relations
	Location string
}

type CreateChoreResult struct {
	StatusCode      int
	SuccessResponse *CreateChoreResponse
	ErrorResponse   *ErrorResponse
}

type GetChoreResponse struct {
	Chore   Chore
	Actions Relations
}

type GetChoreResult struct {
	StatusCode      int
	SuccessResponse *GetChoreResponse
	ErrorResponse   *ErrorResponse
}

type EditChoreParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Deadline    string `json:"deadline"`
}

type EditChoreResponse struct {
	Chore   Chore
	Actions Relations
}

type EditChoreResult struct {
	StatusCode      int
	SuccessResponse *EditChoreResponse
	ErrorResponse   *ErrorResponse
}

type CompleteChoreParams struct {
	CompletedOn string `json:"completed_on"`
}

type CompleteChoreResponse struct {
	Chore   Chore
	Actions Relations
}

type CompleteChoreResult struct {
	StatusCode      int
	SuccessResponse *CompleteChoreResponse
	ErrorResponse   *ErrorResponse
}

type CorrectChoreCompletionParams struct {
	CompletedOn string `json:"completed_on"`
}

type CorrectChoreCompletionResponse struct {
	Chore   Chore
	Actions Relations
}

type CorrectChoreCompletionResult struct {
	StatusCode      int
	SuccessResponse *CorrectChoreCompletionResponse
	ErrorResponse   *ErrorResponse
}

type DeleteChoreResult struct {
	StatusCode    int
	ErrorResponse *ErrorResponse
}

type ChoreTemplate struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	DeactivatedAt *string   `json:"deactivated_at"`
	Links         Relations `json:"_links"`
}

type ChoreTemplateCollection struct {
	Items   []ChoreTemplate `json:"items"`
	More    bool            `json:"more"`
	Links   Relations       `json:"_links"`
	Actions Relations       `json:"_actions"`
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

type ChoreTemplatePickerItem struct {
	Id    int       `json:"id"`
	Name  string    `json:"name"`
	Links Relations `json:"_links"`
}

type ChoreTemplatePickerCollection struct {
	Items []ChoreTemplatePickerItem `json:"items"`
	More  bool                      `json:"more"`
	Links Relations                 `json:"_links"`
}

type BrowseChoreTemplatePickerParams struct {
	Href   string
	Search string
	Limit  int
}

type BrowseChoreTemplatePickerResult struct {
	Collection ChoreTemplatePickerCollection
}

type CreateChoreTemplateParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateChoreTemplateResponse struct {
	ChoreTemplate ChoreTemplate
	Actions       Relations
	Location      string
}

type CreateChoreTemplateResult struct {
	StatusCode      int
	SuccessResponse *CreateChoreTemplateResponse
	ErrorResponse   *ErrorResponse
}

type GetChoreTemplateResponse struct {
	ChoreTemplate ChoreTemplate
	Actions       Relations
}

type GetChoreTemplateResult struct {
	StatusCode      int
	SuccessResponse *GetChoreTemplateResponse
	ErrorResponse   *ErrorResponse
}

type EditChoreTemplateParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type EditChoreTemplateResponse struct {
	ChoreTemplate ChoreTemplate
	Actions       Relations
}

type EditChoreTemplateResult struct {
	StatusCode      int
	SuccessResponse *EditChoreTemplateResponse
	ErrorResponse   *ErrorResponse
}

type DeactivateChoreTemplateResult struct {
	ChoreTemplate ChoreTemplate
	Actions       Relations
}

type ErrorResponse struct {
	Error   string    `json:"error"`
	Links   Relations `json:"_links"`
	Actions Relations `json:"_actions"`
}

type choreTemplateRepresentation struct {
	ChoreTemplate
	Actions Relations `json:"_actions"`
}

type choreRepresentation struct {
	Chore
	Actions Relations `json:"_actions"`
}
