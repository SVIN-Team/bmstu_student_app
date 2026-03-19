package gormmodels

import "github.com/google/uuid"

type Group struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string     `gorm:"column:name;type:varchar(50);unique;not null"`
	HeadmanID *uuid.UUID `gorm:"column:headman_id;type:uuid"`
}

func (Group) TableName() string {
	return "groups"
}
