package info

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
}
