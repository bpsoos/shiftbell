package demo

type Data struct {
	OneOffChores    []OneOffChore    `json:"one_off_chores"`
	ScheduledChores []ScheduledChore `json:"scheduled_chores"`
}

type OneOffChore struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	DeadlineDays        int    `json:"deadline_days"`
	CompletedDays       *int   `json:"completed_days,omitempty"`
	SaveAsChoreTemplate bool   `json:"save_as_chore_template,omitempty"`
}

type ScheduledChore struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DeadlineDays int    `json:"deadline_days"`
	ScheduleName string `json:"schedule_name"`
	IntervalDays int    `json:"interval_days"`
}
