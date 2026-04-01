package middleware

import (
	"stud_hub/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxWithRequestId := logger.ContextWithRequestID(c.Request.Context(), uuid.New().String())
		c.Request = c.Request.WithContext(ctxWithRequestId)
	}
}
