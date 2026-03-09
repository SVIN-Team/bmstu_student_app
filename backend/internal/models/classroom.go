package models

import "github.com/google/uuid"

type Classroom struct {
	ID   uuid.UUID
	Name string
}
