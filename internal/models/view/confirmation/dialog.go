package confirmation

type Dialog struct {
	Heading      string
	Prompt       string
	Name         string
	Supporting   []string
	ActionHref   string
	ActionMethod string
	ActionLabel  string
	Error        string
	Icon         string
}
