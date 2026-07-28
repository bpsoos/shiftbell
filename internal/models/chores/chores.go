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
	Search string
	Offset int
	Limit  int
}

type Chore struct {
	Id            int
	ScheduleId    *int
	PredecessorId *int
	Name          string
	Status        ChoreStatus
	Description   string
	Deadline      time.Time
	CompletedOn   time.Time
}

type ChorePage = GetChoreBatchResult

type ChoreDetails = Chore

type CreateOneOffChoreParams struct {
	Name        string
	Description string
	Deadline    time.Time
}

type CreateChoreParams struct {
	Name                string
	Description         string
	Deadline            time.Time
	ChoreTemplateId     *int
	ScheduleName        string
	IntervalDays        *int
	SaveAsChoreTemplate bool
}

type CreateManualOneOffParams struct {
	Name                string
	Description         string
	Deadline            time.Time
	SaveAsChoreTemplate bool
}

type CreateManualScheduledParams struct {
	Name         string
	Description  string
	Deadline     time.Time
	ScheduleName string
	IntervalDays int
}

type CreateTemplateScheduledParams struct {
	ChoreTemplateId int
	Deadline        time.Time
	ScheduleName    string
	IntervalDays    int
}

type CreateChoreResult struct {
	Chore         *Chore
	ChoreTemplate *choretemplatemodels.ChoreTemplate
}

type EditChoreParams struct {
	Id                      int
	ScheduleId              *int
	Name                    string
	Description             string
	Deadline                time.Time
	AlsoUpdateChoreTemplate bool
}

type EditOneOffChoreParams struct {
	Id          int
	Name        string
	Description string
	Deadline    time.Time
}

type EditScheduledChoreParams struct {
	Id                      int
	Name                    string
	Description             string
	AlsoUpdateChoreTemplate bool
}

type EditChoreResult = ChoreDetails

type CompleteChoreParams struct {
	Id          int
	CompletedOn time.Time
}

type CompleteChoreResult struct {
	Chore     *Chore
	Successor *Chore
}

type CorrectCompletionParams struct {
	Id          int
	CompletedOn time.Time
}

type CorrectCompletionResult struct {
	Chore     *Chore
	Successor *Chore
}
