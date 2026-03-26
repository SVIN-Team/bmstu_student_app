package gormmodels

import (
	"strings"

	"github.com/google/uuid"
)

type Teacher struct {
	ID         uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	FirstName  string    `gorm:"column:first_name;type:varchar(100);not null"`
	LastName   string    `gorm:"column:last_name;type:varchar(100);not null"`
	Patronymic string    `gorm:"column:patronymic;type:varchar(100)"`
}

func (t *Teacher) Name() string {
	parts := make([]string, 0, 3)
	if t.FirstName != "" {
		parts = append(parts, t.FirstName)
	}
	if t.LastName != "" {
		parts = append(parts, t.LastName)
	}
	if t.Patronymic != "" {
		parts = append(parts, t.Patronymic)
	}
	return strings.Join(parts, " ")
}

func (Teacher) TableName() string {
	return "teachers"
}
