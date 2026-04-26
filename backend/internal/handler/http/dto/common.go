package dto

// Common response types used across multiple handlers

// GroupResponse represents group information
type GroupResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SubjectResponse represents subject information
type SubjectResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TeacherResponse represents teacher information
type TeacherResponse struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
}

// RoomResponse represents room/classroom information
type RoomResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
