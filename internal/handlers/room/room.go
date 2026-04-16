package room

import (
	"booking/internal/request"
	"booking/internal/response"
	roomusecase "booking/internal/usecase/room"
	"booking/internal/utils/customErrors"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	log *slog.Logger
	uc  *roomusecase.Usecase
}

func New(log *slog.Logger, uc *roomusecase.Usecase) *Handler {
	return &Handler{log: log, uc: uc}
}

func (h *Handler) CreateRoom() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request.CreateRoomRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.MakeError(
				response.ErrCodeInvalidRequest,
				"invalid request",
			))
			return
		}

		room, err := h.uc.CreateRoom(c.Request.Context(), req.Name, req.Description, req.Capacity)
		if errors.Is(err, customErrors.ErrInvalidData) {
			c.JSON(http.StatusBadRequest, response.MakeError(
				response.ErrCodeInvalidRequest,
				"invalid request",
			))
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.MakeError(
				response.ErrCodeInternalServerError,
				"internal server error",
			))
			return
		}

		c.JSON(http.StatusCreated, gin.H{"room": room})
	}
}

func (h *Handler) GetRooms() gin.HandlerFunc {
	return func(c *gin.Context) {
		rooms, err := h.uc.GetRooms(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.MakeError(
				response.ErrCodeInternalServerError,
				"internal server error",
			))
			return
		}

		c.JSON(http.StatusOK, gin.H{"rooms": rooms})
	}
}
