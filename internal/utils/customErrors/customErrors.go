package customErrors

import "errors"

var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidData        = errors.New("invalid data")
	ErrRoomNotFound       = errors.New("room not found")
	ErrScheduleExists     = errors.New("schedule already exists")
	ErrSlotNotFound       = errors.New("slot not found")
	ErrSlotAlreadyBooked  = errors.New("slot already booked")
	ErrBookingNotFound    = errors.New("booking not found")
	ErrForbidden          = errors.New("forbidden")
)
