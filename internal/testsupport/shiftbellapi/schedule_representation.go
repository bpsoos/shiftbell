package shiftbellapi

type ScheduleRepresentation struct {
	Schedule
	Actions Relations `json:"_actions"`
}
