package dto

import (
	"stud_hub/internal/models"

	"github.com/google/uuid"
)

type SignUpRequest struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	GroupName string `json:"group_name" binding:"required"`
}

func (a *SignUpRequest) ToModel(groupId uuid.UUID) models.User {
	return models.User{
		Email:        a.Email,
		PasswordHash: a.Password,
		FirstName:    a.FirstName,
		LastName:     a.LastName,
		GroupID:      groupId,
	}
}
