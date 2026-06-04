package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, envelope{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, envelope{Success: true, Message: message, Data: data})
}

func BadRequest(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, envelope{Success: false, Error: err})
}

func Unauthorized(c *gin.Context, err string) {
	c.JSON(http.StatusUnauthorized, envelope{Success: false, Error: err})
}

func NotFound(c *gin.Context, err string) {
	c.JSON(http.StatusNotFound, envelope{Success: false, Error: err})
}

func Conflict(c *gin.Context, err string) {
	c.JSON(http.StatusConflict, envelope{Success: false, Error: err})
}

func InternalError(c *gin.Context, err string) {
	c.JSON(http.StatusInternalServerError, envelope{Success: false, Error: err})
}
