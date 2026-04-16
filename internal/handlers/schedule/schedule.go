package schedule

import (
	"booking/internal/request"
	"booking/internal/response"
	scheduleusecase "booking/internal/usecase/schedule"
	"booking/internal/utils/customErrors"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	log *slog.Logger
	uc  *scheduleusecase.Usecase
}

func New(log *slog.Logger, uc *scheduleusecase.Usecase) *Handler {
	return &Handler{log: log, uc: uc}
}

func (h *Handler) CreateSchedule() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("roomId")

		var req request.CreateScheduleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.MakeError(response.ErrCodeInvalidRequest, "invalid request"))
			return
		}

		schedule, err := h.uc.CreateSchedule(c.Request.Context(), roomID, req.DaysOfWeek, req.StartTime, req.EndTime)
		if err != nil {
			switch {
			case errors.Is(err, customErrors.ErrInvalidData):
				c.JSON(http.StatusBadRequest, response.MakeError(response.ErrCodeInvalidRequest, "invalid request"))
			case errors.Is(err, customErrors.ErrRoomNotFound):
				c.JSON(http.StatusNotFound, response.MakeError(response.ErrCodeRoomNotFound, "room not found"))
			case errors.Is(err, customErrors.ErrScheduleExists):
				c.JSON(http.StatusConflict, response.MakeError(response.ErrCodeScheduleExist, "schedule for this room already exists and cannot be changed"))
			default:
				c.JSON(http.StatusInternalServerError, response.MakeError(response.ErrCodeInternalServerError, "internal server error"))
			}
			return
		}

		c.JSON(http.StatusCreated, gin.H{"schedule": schedule})
	}
}
