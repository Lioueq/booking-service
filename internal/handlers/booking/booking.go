package booking

import (
	"booking/internal/request"
	"booking/internal/response"
	bookingusecase "booking/internal/usecase/booking"
	"booking/internal/utils/customErrors"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	log *slog.Logger
	uc  *bookingusecase.Usecase
}

func New(log *slog.Logger, uc *bookingusecase.Usecase) *Handler {
	return &Handler{log: log, uc: uc}
}

func (h *Handler) CreateBooking() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request.CreateBookingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.MakeError(response.ErrCodeInvalidRequest, "invalid request"))
			return
		}

		userID, _ := c.Get("user_id")

		booking, err := h.uc.CreateBooking(c.Request.Context(), toString(userID), req.SlotID, req.CreateConferenceLink)
		if err != nil {
			switch {
			case errors.Is(err, customErrors.ErrInvalidData):
				c.JSON(http.StatusBadRequest, response.MakeError(response.ErrCodeInvalidRequest, "invalid request"))
			case errors.Is(err, customErrors.ErrSlotNotFound):
				c.JSON(http.StatusNotFound, response.MakeError(response.ErrCodeSlotNotFound, "slot not found"))
			case errors.Is(err, customErrors.ErrSlotAlreadyBooked):
				c.JSON(http.StatusConflict, response.MakeError(response.ErrCodeSlotAlreadyBooked, "slot is already booked"))
			default:
				c.JSON(http.StatusInternalServerError, response.MakeError(response.ErrCodeInternalServerError, "internal server error"))
			}
			return
		}

		c.JSON(http.StatusCreated, gin.H{"booking": booking})
	}
}

func (h *Handler) CancelBooking() gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("bookingId")
		userID, _ := c.Get("user_id")

		booking, err := h.uc.CancelBooking(c.Request.Context(), toString(userID), bookingID)
		if err != nil {
			switch {
			case errors.Is(err, customErrors.ErrBookingNotFound):
				c.JSON(http.StatusNotFound, response.MakeError(response.ErrCodeBookingNotFound, "booking not found"))
			case errors.Is(err, customErrors.ErrForbidden), errors.Is(err, customErrors.ErrInvalidRole):
				c.JSON(http.StatusForbidden, response.MakeError(response.ErrCodeForbidden, "cannot cancel another user's booking"))
			case errors.Is(err, customErrors.ErrInvalidData):
				c.JSON(http.StatusBadRequest, response.MakeError(response.ErrCodeInvalidRequest, "invalid request"))
			default:
				c.JSON(http.StatusInternalServerError, response.MakeError(response.ErrCodeInternalServerError, "internal server error"))
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"booking": booking})
	}
}

func (h *Handler) ListBookings() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request.ListBookingsRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.MakeError(response.ErrCodeInvalidRequest, "invalid request"))
			return
		}

		if _, ok := c.GetQuery("page"); ok && req.Page < 1 {
			c.JSON(http.StatusBadRequest, response.MakeError(response.ErrCodeInvalidRequest, "invalid request"))
			return
		}

		if _, ok := c.GetQuery("pageSize"); ok && (req.PageSize < 1 || req.PageSize > 100) {
			c.JSON(http.StatusBadRequest, response.MakeError(response.ErrCodeInvalidRequest, "invalid request"))
			return
		}

		bookings, total, err := h.uc.ListBookings(c.Request.Context(), req.Page, req.PageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.MakeError(response.ErrCodeInternalServerError, "internal server error"))
			return
		}

		page := req.Page
		if page <= 0 {
			page = 1
		}
		pageSize := req.PageSize
		if pageSize <= 0 {
			pageSize = 20
		}
		if pageSize > 100 {
			pageSize = 100
		}

		c.JSON(http.StatusOK, gin.H{
			"bookings": bookings,
			"pagination": gin.H{
				"page":     page,
				"pageSize": pageSize,
				"total":    total,
			},
		})
	}
}

func (h *Handler) MyBookings() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		bookings, err := h.uc.MyBookings(c.Request.Context(), toString(userID))
		if err != nil {
			switch {
			case errors.Is(err, customErrors.ErrInvalidRole):
				c.JSON(http.StatusForbidden, response.MakeError(response.ErrCodeForbidden, "forbidden"))
			default:
				c.JSON(http.StatusInternalServerError, response.MakeError(response.ErrCodeInternalServerError, "internal server error"))
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"bookings": bookings})
	}
}

func toString(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
