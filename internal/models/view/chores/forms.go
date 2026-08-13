package chores

type Field struct {
	Value string
	Error string
}

type EditForm struct {
	ActionHref              string
	CancelHref              string
	Scheduled               bool
	Name                    Field
	Description             Field
	Deadline                Field
	AlsoUpdateChoreTemplate bool
	SummaryError            string
}

type ManualOneOffForm struct {
	ActionHref     string
	CancelHref     string
	SummaryError   string
	Submitted      bool
	Name           Field
	Description    Field
	Deadline       Field
	SaveAsTemplate bool
}

func (form ManualOneOffForm) NeedsDefaultDeadline() bool {
	return !form.Submitted && form.Deadline.Value == ""
}

type TemplateOneOffForm struct {
	ActionHref               string
	CancelHref               string
	SummaryError             string
	Submitted                bool
	ChoreTemplateId          int
	ChoreTemplateName        string
	ChoreTemplateDescription string
	Deadline                 Field
}

func (form TemplateOneOffForm) HasSelectedTemplate() bool {
	return form.ChoreTemplateName != "" || form.ChoreTemplateId > 0
}

func (form TemplateOneOffForm) NeedsDefaultDeadline() bool {
	return !form.Submitted && form.Deadline.Value == ""
}
