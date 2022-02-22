package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RequestID = "x-request-id"

// RequestIDMiddleware provides an ID to the request for traceback purposes.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(RequestID, uuid.New().String())
		c.Next()
	}
}

// ResponseMiddleware sets the header with the request header.
func ResponseIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestUUID string
		val, ok := c.Get(RequestID)
		if !ok {
			zap.L().Error("response without request-id")
			requestUUID = uuid.New().String()
		} else {
			requestUUID = val.(string)
		}
		c.Writer.Header().Set(RequestID, requestUUID)
		c.Next()
	}
}
