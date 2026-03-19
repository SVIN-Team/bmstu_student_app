package gormmodels

import "github.com/google/uuid"

type Group struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string     `gorm:"column:name;type:varchar(50);unique;uniqueIndex:groups_name_key;not null"`
	HeadmanID *uuid.UUID `gorm:"column:headman_id;type:uuid"`

	Headman   *User      `gorm:"foreignKey:HeadmanID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Users     []User     `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Lessons   []Lesson   `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Queues    []Queue    `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Group) TableName() string {
	return "groups"
}
