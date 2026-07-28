package chores

import (
	"context"
	"time"

	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
)

type GetChoreBatchResult struct {
	Chores []Chore
	More   bool
}

type NewChoreCtxValues struct {
	SelectedChoreTemplate *choretemplatemodels.ChoreTemplate
	IsManual              bool
}

type contextKey string

var newChoreCtxKey contextKey = "newChore"

func WithNewChoreCtxValues(ctx context.Context, values *NewChoreCtxValues) context.Context {
	return context.WithValue(ctx, newChoreCtxKey, values)
}

func GetSelectedChoreTemplate(ctx context.Context) *choretemplatemodels.ChoreTemplate {
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

type BrowseChoresParams struct {
	Status ChoreStatus
	Offset int
	Limit  int
}

type Chore struct {
	Id          int
	Name        string
	Status      ChoreStatus
	Description string
	Deadline    time.Time
	CompletedOn time.Time
}

type ChorePage = GetChoreBatchResult

type CreateChoreParams struct {
	Name        string
	Description string
	Deadline    time.Time
}

type CreateChoreInput struct {
	Name                string
	Description         string
	Deadline            time.Time
	ChoreTemplateId     *int
	ScheduleName        string
	IntervalDays        *int
	SaveAsChoreTemplate bool
}

type CreateManualOneOffInput struct {
	Name                string
	Description         string
	Deadline            time.Time
	SaveAsChoreTemplate bool
}

type CreateManualScheduledInput struct {
	Name         string
	Description  string
	Deadline     time.Time
	ScheduleName string
	IntervalDays int
}

type CreateTemplateScheduledInput struct {
	ChoreTemplateId int
	Deadline        time.Time
	ScheduleName    string
	IntervalDays    int
}

type CreateChoreResult struct {
	Chore         *Chore
	ChoreTemplate *choretemplatemodels.ChoreTemplate
}
