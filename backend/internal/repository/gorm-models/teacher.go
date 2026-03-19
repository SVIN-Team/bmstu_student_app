package gormmodels

import "github.com/google/uuid"

type Teacher struct {
	ID       uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	FullName string    `gorm:"column:full_name;type:varchar(255);not null"`
}

func (Teacher) TableName() string {
	return "teachers"
}
