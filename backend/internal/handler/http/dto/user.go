package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserResponse represents the response body for a user
type UserResponse struct {
	ID        string         `json:"id"`
	Email     string         `json:"email"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Role      string         `json:"role"`
	Group     *GroupResponse `json:"group,omitempty"`
	IsBlocked *bool          `json:"is_blocked,omitempty"`
	CreatedAt *time.Time     `json:"created_at,omitempty"`
}

// UserMinimalResponse represents minimal user information
type UserMinimalResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UpdateUserRequest represents the request body for updating user profile
type UpdateUserRequest struct {
	FirstName *string    `json:"first_name"`
	LastName  *string    `json:"last_name"`
	GroupID   *uuid.UUID `json:"group_id"`
}

// TransferHeadmanRoleRequest represents the request for transferring headman role
type TransferHeadmanRoleRequest struct {
	ToUserID uuid.UUID `json:"to_user_id" binding:"required"`
}

// TransferHeadmanRoleResponse represents the response for transferring headman role
type TransferHeadmanRoleResponse struct {
	PreviousHeadmanID string `json:"previous_headman_id"`
	NewHeadmanID      string `json:"new_headman_id"`
	GroupID           string `json:"group_id"`
}
