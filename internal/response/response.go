package response

type ErrorCode string

const (
	ErrCodeInvalidRequest      ErrorCode = "INVALID_REQUEST"
	ErrCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrCodeRoomNotFound        ErrorCode = "ROOM_NOT_FOUND"
	ErrCodeSlotNotFound        ErrorCode = "SLOT_NOT_FOUND"
	ErrCodeBookingNotFound     ErrorCode = "BOOKING_NOT_FOUND"
	ErrCodeInternalServerError ErrorCode = "INTERNAL_ERROR"
	ErrCodeScheduleExist       ErrorCode = "SCHEDULE_EXISTS"
	ErrCodeSlotAlreadyBooked   ErrorCode = "SLOT_ALREADY_BOOKED"
	ErrCodeNotFound            ErrorCode = "NOT_FOUND"
	ErrCodeForbidden           ErrorCode = "FORBIDDEN"
)

type ErrorResponse struct {
	Error Error `json:"error"`
}

type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func MakeError(code ErrorCode, message string) ErrorResponse {
	return ErrorResponse{
		Error: Error{
			Code:    code,
			Message: message,
		},
	}
}
