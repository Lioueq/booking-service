package domain

import "time"

type Room struct {
	RoomID      string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Capacity    int       `json:"capacity"`
	CreatedAt   time.Time `json:"createdAt"`
}
