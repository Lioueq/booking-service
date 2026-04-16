package slot

import (
	"booking/internal/request"
	"booking/internal/response"
	slotusecase "booking/internal/usecase/slot"
	"booking/internal/utils/customErrors"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	log *slog.Logger
	uc  *slotusecase.Usecase
}

func New(log *slog.Logger, uc *slotusecase.Usecase) *Handler {
	return &Handler{log: log, uc: uc}
}

func (h *Handler) ListSlots() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("roomId")

		var req request.ListSlotsRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.MakeError(response.ErrCodeInvalidRequest, "invalid request"))
			return
		}

		slots, err := h.uc.GetAvailableSlots(c.Request.Context(), roomID, req.Date)
		if err != nil {
			switch {
			case errors.Is(err, customErrors.ErrInvalidData):
				c.JSON(http.StatusBadRequest, response.MakeError(response.ErrCodeInvalidRequest, "invalid request"))
			case errors.Is(err, customErrors.ErrRoomNotFound):
				c.JSON(http.StatusNotFound, response.MakeError(response.ErrCodeRoomNotFound, "room not found"))
			default:
				c.JSON(http.StatusInternalServerError, response.MakeError(response.ErrCodeInternalServerError, "internal server error"))
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"slots": slots})
	}
}
