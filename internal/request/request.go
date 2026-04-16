package request

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type DummyLoginRequest struct {
	Role string `json:"role" binding:"required"`
}

type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"`
}

type CreateScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek" binding:"required"`
	StartTime  string `json:"startTime" binding:"required"`
	EndTime    string `json:"endTime" binding:"required"`
}

type ListSlotsRequest struct {
	Date string `form:"date" binding:"required"`
}

type CreateBookingRequest struct {
	SlotID               string `json:"slotId" binding:"required"`
	CreateConferenceLink bool   `json:"createConferenceLink"`
}

type ListBookingsRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}
