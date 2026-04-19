package http

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// HTTP response helper functions to avoid code duplication

// SuccessResponse returns a success response with data
func SuccessResponse(ctx *gin.Context, statusCode int, data interface{}) {
    ctx.JSON(statusCode, gin.H{
        "success": true,
        "data":    data,
    })
}

// SuccessResponseWithPagination returns a success response with data and pagination
func SuccessResponseWithPagination(ctx *gin.Context, statusCode int, data interface{}, page, perPage, total int) {
    ctx.JSON(statusCode, gin.H{
        "success": true,
        "data":    data,
        "pagination": gin.H{
            "page":     page,
            "per_page": perPage,
            "total":    total,
        },
    })
}

// ErrorResponse returns an error response
func ErrorResponse(ctx *gin.Context, statusCode int, code, message string) {
    ctx.JSON(statusCode, gin.H{
        "success": false,
        "error": gin.H{
            "code":    code,
            "message": message,
        },
    })
}

// Common error responses

func ValidationError(ctx *gin.Context, message string) {
    ErrorResponse(ctx, http.StatusBadRequest, "VALIDATION_ERROR", message)
}

func UnauthorizedError(ctx *gin.Context, message string) {
    ErrorResponse(ctx, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func ForbiddenError(ctx *gin.Context, message string) {
    ErrorResponse(ctx, http.StatusForbidden, "FORBIDDEN", message)
}

func NotFoundError(ctx *gin.Context, message string) {
    ErrorResponse(ctx, http.StatusNotFound, "NOT_FOUND", message)
}

func ConflictError(ctx *gin.Context, code, message string) {
    ErrorResponse(ctx, http.StatusConflict, code, message)
}

func InternalError(ctx *gin.Context, message string) {
    ErrorResponse(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// Specific queue errors

func QueueFullError(ctx *gin.Context) {
    ErrorResponse(ctx, http.StatusConflict, "QUEUE_FULL", "Queue is full")
}

func QueueClosedError(ctx *gin.Context) {
    ErrorResponse(ctx, http.StatusConflict, "QUEUE_CLOSED", "Queue is closed for sign-ups")
}

func QueueNotOpenError(ctx *gin.Context) {
    ErrorResponse(ctx, http.StatusConflict, "QUEUE_NOT_OPEN", "Queue is not open for sign-ups")
}

func AlreadyExistsError(ctx *gin.Context, message string) {
    ErrorResponse(ctx, http.StatusConflict, "ALREADY_EXISTS", message)
}

func AlreadyInQueueError(ctx *gin.Context) {
    ErrorResponse(ctx, http.StatusConflict, "ALREADY_EXISTS", "You are already signed up for this queue")
}
