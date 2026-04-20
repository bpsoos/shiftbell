package models

import "time"

type GetChoreBatchResult struct {
	Chores []Chore
	More   bool
}

type ChoreStatus string

const (
	ChoreStatusComplete   ChoreStatus = "complete"
	ChoreStatusIncomplete ChoreStatus = "incomplete"
)

type Chore struct {
	Id              int
	Status          ChoreStatus
	Description     string
	IntervalDays    int
	Deadline        time.Time
	LastCompletedAt time.Time
	CompletedAt     time.Time
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
