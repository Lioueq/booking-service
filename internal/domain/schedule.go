package domain

import "time"

type Schedule struct {
	ScheduleID string    `json:"id"`
	RoomID     string    `json:"roomId"`
	DaysOfWeek []int     `json:"daysOfWeek"`
	StartTime  string    `json:"startTime"`
	EndTime    string    `json:"endTime"`
	CreatedAt  time.Time `json:"createdAt"`
}
