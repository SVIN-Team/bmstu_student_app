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

// GetLessons godoc
// @Summary List lessons
// @Description Returns lessons by group and date range.
// @Tags Lessons
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param group_id query string false "Group ID (UUID). If omitted, current user's group is used."
// @Param date_from query string true "Start date in YYYY-MM-DD format"
// @Param date_to query string true "End date in YYYY-MM-DD format"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.LessonResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /lessons [get]
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

// GetLessonByID godoc
// @Summary Get lesson by ID
// @Description Returns lesson details by lesson ID.
// @Tags Lessons
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param id path string true "Lesson ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.LessonResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /lessons/{id} [get]
func (h *ScheduleHandler) GetLessonByID(ctx *gin.Context) {
	lessonIDStr := ctx.Param("id")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		ValidationError(ctx, "Invalid lesson ID format")
		return
	}

	lesson, err := h.scheduleUseCase.GetLessonByID(ctx, lessonID)
	if errors.Is(err, errors2.ErrLessonNotFound) {
		NotFoundError(ctx, "Lesson not found")
		return
	} else if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to get lesson: %v", err))
		logger.Errorf(ctx, "Failed to get lesson: %v", err)
		return
	}

	SuccessResponse(ctx, http.StatusOK, h.lessonToDTO(lesson))
}

// ImportSchedule godoc
// @Summary Import schedule
// @Description Imports lessons from uploaded JSON file (admin endpoint when enabled in routing).
// @Tags Lessons
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param file formData file true "JSON file with import payload"
// @Success 200 {object} dto.SuccessResponse{data=dto.ImportScheduleResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /lessons/imports [post]
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
