package gormmodels

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string     `gorm:"column:email;type:varchar(255);not null;unique;uniqueIndex:users_email_key"`
	PasswordHash *string    `gorm:"column:password_hash;type:varchar(60)"`
	FirstName    string     `gorm:"column:first_name;type:varchar(100);not null"`
	LastName     string     `gorm:"column:last_name;type:varchar(100);not null"`
	Patronymic   string     `gorm:"column:patronymic;type:varchar(100)"`
	Role         UserRole   `gorm:"column:role;type:user_role;not null;default:'student'"`
	GroupID      *uuid.UUID `gorm:"column:group_id;type:uuid"`
	IsBlocked    bool       `gorm:"column:is_blocked;not null;default:false"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;autoCreateTime"`

	Group        *Group     `gorm:"foreignKey:GroupID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (User) TableName() string {
	return "users"
}
