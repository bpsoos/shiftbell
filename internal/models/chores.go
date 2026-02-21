package models

type ChoreType struct {
	Id           int
	Description  string
	IntervalDays int
}

type GetChoreTypeBatchResult struct {
	ChoreTypes []ChoreType
	More       bool
}
