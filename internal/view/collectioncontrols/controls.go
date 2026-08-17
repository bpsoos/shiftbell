package collectioncontrols

type HiddenInput struct {
	Name  string
	Value string
}

type Model struct {
	ActionHref        string
	AutofocusSearch   bool
	HiddenInputs      []HiddenInput
	Search            string
	SearchInputID     string
	SearchPlaceholder string
}
