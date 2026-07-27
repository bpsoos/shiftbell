package models

import (
	"context"
	"time"
)

type GetChoreBatchResult struct {
	Chores []Chore
	More   bool
}

type NewChoreCtxValues struct {
	SelectedChoreTemplate *ChoreTemplate
	IsManual              bool
}

type contextKey string

var newChoreCtxKey contextKey = "newChore"

func WithNewChoreCtxValues(ctx context.Context, values *NewChoreCtxValues) context.Context {
	return context.WithValue(ctx, newChoreCtxKey, values)
}

func GetSelectedChoreTemplate(ctx context.Context) *ChoreTemplate {
	if newChoreCtxValues, ok := ctx.Value(newChoreCtxKey).(*NewChoreCtxValues); ok {
		return newChoreCtxValues.SelectedChoreTemplate
	}
	return nil
}

func GetIsManual(ctx context.Context) bool {
	if newChoreCtxValues, ok := ctx.Value(newChoreCtxKey).(*NewChoreCtxValues); ok {
		return newChoreCtxValues.IsManual
	}
	return false
}

type NewChoreTemplateComponent string

const (
	NewChoreTemplateComponentInputTypeSelector NewChoreTemplateComponent = "inputTypeSelector"
	NewChoreTemplateComponentBaseInputs        NewChoreTemplateComponent = "baseInputs"
)

type ChoreStatus string

const (
	ChoreStatusActive    ChoreStatus = "active"
	ChoreStatusCompleted ChoreStatus = "completed"
)

type Chore struct {
	Id          int
	Name        string
	Status      ChoreStatus
	Description string
	Deadline    time.Time
	CompletedOn time.Time
}

type CreateChoreParams struct {
	Name        string
	Description string
	Deadline    time.Time
}

type CreateChoreInput struct {
	Name                string
	Description         string
	Deadline            time.Time
	SaveAsChoreTemplate bool
}

type CreateManualOneOffInput struct {
	Name                string
	Description         string
	Deadline            time.Time
	SaveAsChoreTemplate bool
}

type CreateChoreResult struct {
	Chore         *Chore
	ChoreTemplate *ChoreTemplate
}

type CreateChoreTemplateParams struct {
	Name        string
	Description string
}

type EditChoreTemplateParams struct {
	Id          int
	Name        string
	Description string
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
