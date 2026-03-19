package gormmodels

import "github.com/google/uuid"

type Subject struct {
	ID   uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Name string    `gorm:"column:name;type:varchar(200);not null;unique"`
}

func (Subject) TableName() string {
	return "subjects"
}
