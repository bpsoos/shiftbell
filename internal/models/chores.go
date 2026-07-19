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
	SelectedChoreType *ChoreType
	IsManual          bool
}

type contextKey string

var newChoreCtxKey contextKey = "newChore"

func WithNewChoreCtxValues(ctx context.Context, values *NewChoreCtxValues) context.Context {
	return context.WithValue(ctx, newChoreCtxKey, values)
}

func GetSelectedChoreType(ctx context.Context) *ChoreType {
	if newChoreCtxValues, ok := ctx.Value(newChoreCtxKey).(*NewChoreCtxValues); ok {
		return newChoreCtxValues.SelectedChoreType
	}
	return nil
}

func GetIsManual(ctx context.Context) bool {
	if newChoreCtxValues, ok := ctx.Value(newChoreCtxKey).(*NewChoreCtxValues); ok {
		return newChoreCtxValues.IsManual
	}
	return false
}

type NewChoreTypeComponent string

const (
	NewChoreTypeComponentInputTypeSelector NewChoreTypeComponent = "inputTypeSelector"
	NewChoreTypeComponentBaseInputs        NewChoreTypeComponent = "baseInputs"
)

type ChoreStatus string

const (
	ChoreStatusComplete   ChoreStatus = "complete"
	ChoreStatusIncomplete ChoreStatus = "incomplete"
)

type Chore struct {
	Id          int
	Status      ChoreStatus
	Description string
	Deadline    time.Time
	CompletedAt time.Time
}

type CreateChoreParams struct {
	Name        string
	Description string
	Deadline    time.Time
}

type ChoreType struct {
	Id          int
	Description string
	Name        string
}

type GetChoreTypeBatchResult struct {
	ChoreTypes []ChoreType
	More       bool
}
