package domain

import "time"

type Booking struct {
	BookingID      string    `json:"id"`
	SlotID         string    `json:"slotId"`
	UserID         string    `json:"userId"`
	Status         string    `json:"status"`
	ConferenceLink string    `json:"conferenceLink"`
	CreatedAt      time.Time `json:"createdAt"`
}
