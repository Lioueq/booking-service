package auth

import (
	"booking/internal/logger"
	"booking/internal/request"
	"booking/internal/response"
	authusecase "booking/internal/usecase/auth"
	"booking/internal/utils/customErrors"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	log *slog.Logger
	uc  *authusecase.Usecase
}

func New(log *slog.Logger, uc *authusecase.Usecase) *Handler {
	return &Handler{log: log, uc: uc}
}

func (h *Handler) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.MakeError(
				response.ErrCodeInvalidRequest,
				"invalid request",
			))
			return
		}

		user, err := h.uc.Register(c.Request.Context(), req.Email, req.Password, req.Role)
		if err != nil {
			switch {
			case errors.Is(err, customErrors.ErrInvalidRole):
				c.JSON(http.StatusBadRequest, response.MakeError(
					response.ErrCodeInvalidRequest,
					"invalid request",
				))
			case errors.Is(err, customErrors.ErrInvalidData):
				c.JSON(http.StatusBadRequest, response.MakeError(
					response.ErrCodeInvalidRequest,
					"invalid request",
				))
			default:
				c.JSON(http.StatusInternalServerError, response.MakeError(
					response.ErrCodeInternalServerError,
					"internal server error",
				))
			}
			return
		}

		c.JSON(http.StatusCreated, gin.H{"user": user})
	}
}

func (h *Handler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.MakeError(
				response.ErrCodeInvalidRequest,
				"invalid request",
			))
			return
		}

		token, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, customErrors.ErrInvalidCredentials):
				c.JSON(http.StatusUnauthorized, response.MakeError(
					response.ErrCodeUnauthorized,
					"unauthorized",
				))
			default:
				h.log.Error("login failed", logger.Err(err))
				c.JSON(http.StatusInternalServerError, response.MakeError(
					response.ErrCodeInternalServerError,
					"internal server error",
				))
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func (h *Handler) DummyLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request.DummyLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.MakeError(
				response.ErrCodeInvalidRequest,
				"invalid request",
			))
			return
		}

		token, err := h.uc.DummyLogin(req.Role)
		if err != nil {
			switch {
			case errors.Is(err, customErrors.ErrInvalidRole):
				c.JSON(http.StatusBadRequest, response.MakeError(
					response.ErrCodeInvalidRequest,
					"invalid request",
				))
			default:
				h.log.Error("dummy login failed", logger.Err(err))
				c.JSON(http.StatusInternalServerError, response.MakeError(
					response.ErrCodeInternalServerError,
					"internal server error",
				))
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}
