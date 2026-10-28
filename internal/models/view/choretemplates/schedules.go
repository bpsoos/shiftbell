package choretemplates

type ScheduleList struct {
	Heading      string
	TemplateName string
	TemplateHref string
	EmptyMessage string
	Error        string
	Notice       string
	State        string
	Total        int
	Inline       bool
	ViewAllHref  string
	Schedules    []ScheduleLink
}

type ScheduleLink struct {
	Name string
	Href string
}
