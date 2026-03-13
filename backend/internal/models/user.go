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
	Id         uuid.UUID
	Email      string
	PswHash    string
	FirstName  string
	LastName   string
	Patronymic string
	Role       RoleType
	GroupId    uuid.UUID
	CreatedAt  time.Time
}
