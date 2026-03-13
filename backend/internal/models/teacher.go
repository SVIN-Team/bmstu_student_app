package models

import "github.com/google/uuid"

type Teacher struct {
	Id         uuid.UUID
	FirstName  string
	LastName   string
	Patronymic string
}
