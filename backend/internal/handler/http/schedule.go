package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	errors2 "stud_hub/internal/errors"
	"stud_hub/internal/handler/http/dto"
	"stud_hub/internal/models"
	"stud_hub/internal/usecase"
	"stud_hub/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ScheduleUseCase interface {
	GetSchedule(ctx context.Context, groupID uuid.UUID, from, to time.Time) ([]models.LessonDetails, error)
	GetLessonByID(ctx context.Context, lessonID uuid.UUID) (models.LessonDetails, error)
	ImportSchedule(ctx context.Context, rows []models.ScheduleImportRow, replaceExisting bool) (usecase.ImportResult, error)
}

type ScheduleHandler struct {
	scheduleUseCase ScheduleUseCase
}

func NewScheduleHandler(scheduleUseCase ScheduleUseCase) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleUseCase: scheduleUseCase,
	}
}

// GetLessons returns lessons filtered by group and date range
// GET /lessons?group_id=xxx&date_from=2026-02-23&date_to=2026-02-28
func (h *ScheduleHandler) GetLessons(ctx *gin.Context) {
	groupIDStr := ctx.Query("group_id")
	dateFromStr := ctx.Query("date_from")
	dateToStr := ctx.Query("date_to")

	if dateFromStr == "" || dateToStr == "" {
		ValidationError(ctx, "date_from and date_to are required")
		return
	}

	// If no group_id provided, use the user's group
	var groupID uuid.UUID
	if groupIDStr == "" {
		userGroupID, exists := ctx.Get("user_group_id")
		if !exists {
			ValidationError(ctx, "group_id is required")
			return
		}
		groupID = userGroupID.(uuid.UUID)
	} else {
		var err error
		groupID, err = uuid.Parse(groupIDStr)
		if err != nil {
			ValidationError(ctx, "Invalid group_id format")
			return
		}
	}

	dateFrom, err := time.Parse("2006-01-02", dateFromStr)
	if err != nil {
		ValidationError(ctx, "Invalid date_from format. Use YYYY-MM-DD")
		return
	}

	dateTo, err := time.Parse("2006-01-02", dateToStr)
	if err != nil {
		ValidationError(ctx, "Invalid date_to format. Use YYYY-MM-DD")
		return
	}

	lessons, err := h.scheduleUseCase.GetSchedule(ctx, groupID, dateFrom, dateTo)
	if errors.Is(err, errors2.ErrGroupNotFound) {
		NotFoundError(ctx, "Group not found")
		return
	} else if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to get lessons: %v", err))
		logger.Errorf(ctx, "Failed to get lessons: %v", err)
		return
	}

	response := make([]dto.LessonResponse, len(lessons))
	for i, lesson := range lessons {
		response[i] = h.lessonToDTO(lesson)
	}

	SuccessResponse(ctx, http.StatusOK, response)
}

// GetLessonByID returns a single lesson by ID
// GET /lessons/:id
func (h *ScheduleHandler) GetLessonByID(ctx *gin.Context) {
	lessonIDStr := ctx.Param("id")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid lesson ID format",
			},
		})
		return
	}

	lesson, err := h.scheduleUseCase.GetLessonByID(ctx, lessonID)
	if errors.Is(err, errors2.ErrLessonNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Lesson not found",
			},
		})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": fmt.Sprintf("Failed to get lesson: %v", err),
			},
		})
		logger.Errorf(ctx, "Failed to get lesson: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    h.lessonToDTO(lesson),
	})
}

// ImportSchedule imports schedule from JSON file
// POST /lessons/imports
func (h *ScheduleHandler) ImportSchedule(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ValidationError(ctx, "file is required")
		return
	}
	defer file.Close()

	// Check file extension
	filename := header.Filename
	if !strings.HasSuffix(strings.ToLower(filename), ".json") {
		ValidationError(ctx, "Only JSON files are supported")
		return
	}

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		InternalError(ctx, "Failed to read file")
		logger.Errorf(ctx, "Failed to read file: %v", err)
		return
	}

	// Parse JSON
	var req dto.ImportScheduleRequest
	if err := json.Unmarshal(fileContent, &req); err != nil {
		ValidationError(ctx, fmt.Sprintf("Invalid JSON format: %v", err))
		return
	}

	// Validate
	if err := ctx.ShouldBind(&req); err != nil {
		ValidationError(ctx, fmt.Sprintf("Validation error: %v", err))
		return
	}

	// Convert DTO to model
	rows := make([]models.ScheduleImportRow, 0, len(req.Lessons))
	for i, lessonReq := range req.Lessons {
		// Parse teacher name
		teacherParts := strings.Fields(lessonReq.TeacherName)
		var lastName, firstName, patronymic string
		if len(teacherParts) >= 1 {
			lastName = teacherParts[0]
		}
		if len(teacherParts) >= 2 {
			firstName = teacherParts[1]
		}
		if len(teacherParts) >= 3 {
			patronymic = teacherParts[2]
		}

		row := models.ScheduleImportRow{
			GroupName:         req.GroupID.String(), // Note: This should be group name, but API expects group_id
			SubjectName:       lessonReq.SubjectName,
			TeacherLastName:   lastName,
			TeacherFirstName:  firstName,
			TeacherPatronymic: patronymic,
			RoomName:          lessonReq.RoomName,
			LessonType:        models.LessonType(lessonReq.Type),
			StartsAt:          lessonReq.StartsAt,
			EndsAt:            lessonReq.EndsAt,
			LineNum:           i + 1,
		}
		rows = append(rows, row)
	}

	// Import schedule (replaceExisting defaults to true for admin imports)
	result, err := h.scheduleUseCase.ImportSchedule(ctx, rows, true)
	if err != nil && !errors.Is(err, errors2.ErrImportFailed) {
		InternalError(ctx, fmt.Sprintf("Import failed: %v", err))
		logger.Errorf(ctx, "Failed to import schedule: %v", err)
		return
	}

	// Build response
	response := dto.ImportScheduleResponse{
		ImportedCount: result.SuccessCount,
		Errors:        result.Errors,
	}

	// TODO: Track created subjects and teachers if needed
	// For now, we'll just return empty arrays as the spec shows them as optional

	SuccessResponse(ctx, http.StatusOK, response)
}

// Helper function to convert lesson to DTO
func (h *ScheduleHandler) lessonToDTO(lesson models.LessonDetails) dto.LessonResponse {
	response := dto.LessonResponse{
		ID: lesson.ID.String(),
		Subject: dto.SubjectResponse{
			ID:   lesson.SubjectID.String(),
			Name: lesson.SubjectName,
		},
		Teacher: dto.TeacherResponse{
			ID:       lesson.TeacherID.String(),
			FullName: lesson.TeacherName,
		},
		Type:     string(lesson.LessonType),
		StartsAt: lesson.StartsAt,
		EndsAt:   lesson.EndsAt,
	}

	if lesson.RoomID != uuid.Nil {
		response.Room = &dto.RoomResponse{
			ID:   lesson.RoomID.String(),
			Name: lesson.RoomName,
		}
	}

	if lesson.QueueID != nil {
		queueIDStr := lesson.QueueID.String()
		response.QueueID = &queueIDStr
	}

	return response
}
