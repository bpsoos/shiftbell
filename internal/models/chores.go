package models

import "time"

type GetChoreBatchResult struct {
	Chores []Chore
	More   bool
}

type Chore struct {
	Id           int
	IsCompleted  bool
	Description  string
	IntervalDays int
	Deadline     time.Time
	CreatedAt    time.Time
	CompletedAt  time.Time
}

type ChoreType struct {
	Id           int
	Description  string
	IntervalDays int
}

type GetChoreTypeBatchResult struct {
	ChoreTypes []ChoreType
	More       bool
}
