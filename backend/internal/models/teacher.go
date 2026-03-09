package models

import "github.com/google/uuid"

type Teacher struct {
	ID         uuid.UUID
	FirstName  string
	LastName   string
	Patronymic string
}
