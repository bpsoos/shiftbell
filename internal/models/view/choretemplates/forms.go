package choretemplates

type Field struct {
	Value string
	Error string
}

type EditForm struct {
	ActionHref   string
	CancelHref   string
	Name         Field
	Description  Field
	SummaryError string
}
