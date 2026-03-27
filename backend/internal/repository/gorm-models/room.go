package gormmodels

import "github.com/google/uuid"

type Room struct {
	ID   uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Name *string   `gorm:"column:name;type:varchar(50);unique;uniqueIndex:rooms_name_key"`
}

func (Room) TableName() string {
	return "rooms"
}
