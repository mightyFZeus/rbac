package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"password"`
	Role      string    `json:"role" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Status    string    `json:"status" gorm:"type:varchar(20);default:'pending';check:status IN ('active','pending')"`

	OrganizationID uuid.UUID    `json:"organizationId"`
	Organization   Organization ` json:"-"  gorm:"foreignKey:OrganizationID"`
}

type Admin struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name          string         `json:"name" gorm:"not null"`
	Email         string         `json:"email" gorm:"uniqueIndex;not null"`
	Role          string         `json:"role" gorm:"not null"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	Status        string         `json:"status" gorm:"type:varchar(20);default:'pending';check:status IN ('active','pending')"`
	CreatedBy     uuid.UUID      `json:"createdBy"`
	SuperAdmin    uuid.UUID      ` json:"-"  gorm:"foreignKey:CreatedBy"`
	Organizations []Organization `json:"-" gorm:"foreignKey:AdminID"`
	Password      string         `json:"-" gorm:"not null"`
}

type Organization struct {
	ID          uuid.UUID ` json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Email       string    `json:"email" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AdminID     uuid.UUID `json:"adminId"`
	Admin       Admin     `json:"-" gorm:"foreignKey:AdminID"`

	Users []User ` json:"-"  gorm:"foreignKey:OrganizationID"`
}

type AdminInvites struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AdminId   uuid.UUID `json:"adminId"`
	TokenHash string    `json:"tokenHash" gorm:"not null"`
	ExpiresAt time.Time `json:"expiresAt"`
	UsedAt    time.Time `json:"usedAt"`
	CreatedAt time.Time `json:"createdAt" gorm:"not null"`
}
type UserInvites struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserId    uuid.UUID `json:"userId" gorm:"not null"`
	TokenHash string    `json:"tokenHash" gorm:"not null"`
	ExpiresAt time.Time `json:"expiresAt"`
	UsedAt    time.Time `json:"usedAt"`
	CreatedAt time.Time `json:"createdAt" gorm:"not null"`
	Email     string    `json:"email" gorm:"not null"`
}
