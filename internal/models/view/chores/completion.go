package chores

type CompletionDialog struct {
	Name           string
	ActionHref     string
	CompletedOn    Field
	CompletedOnMax string
}

type CompletionCorrectionForm struct {
	ActionHref     string
	CancelHref     string
	CompletedOn    Field
	CompletedOnMax string
	SummaryError   string
}
