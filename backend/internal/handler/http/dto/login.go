package dto

import (
	"stud_hub/internal/models"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (req *LoginRequest) ToModel() models.User {
	return models.User{
		Email:        req.Email,
		PasswordHash: req.Password,
	}
}
