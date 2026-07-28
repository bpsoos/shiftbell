package choretemplates

import "time"

type CreateChoreTemplateParams struct {
	Name        string
	Description string
}

type EditChoreTemplateParams struct {
	Id          int
	Name        string
	Description string
}

type DeactivateChoreTemplateParams struct {
	Id            int
	DeactivatedAt time.Time
}

type ChoreTemplateFilter string

const (
	ChoreTemplateFilterActive      ChoreTemplateFilter = "active"
	ChoreTemplateFilterDeactivated ChoreTemplateFilter = "deactivated"
)

type BrowseChoreTemplatesParams struct {
	Filter ChoreTemplateFilter
	Search string
	Offset int
	Limit  int
}

type ChoreTemplate struct {
	Id            int
	Description   string
	Name          string
	DeactivatedAt *time.Time
}

type ChoreTemplateDetails struct {
	ChoreTemplate
	ActiveScheduleCount      int
	DeactivatedScheduleCount int
}

type GetChoreTemplateBatchResult struct {
	ChoreTemplates []ChoreTemplate
	More           bool
}

type ChoreTemplatePage = GetChoreTemplateBatchResult
