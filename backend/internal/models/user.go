package models

import (
	"time"

	"github.com/google/uuid"
)

type RoleType string

const (
	RoleAdmin   RoleType = "admin"
	RoleHeadman RoleType = "headman"
	RoleStudent RoleType = "student"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Patronymic   string
	Role         RoleType
	GroupID      uuid.UUID
	CreatedAt    time.Time
}
