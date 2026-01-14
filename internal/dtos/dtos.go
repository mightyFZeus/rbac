package dtos

import (
	"github.com/google/uuid"
)

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CreateUserPayload struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`

	OrganizationID uuid.UUID `json:"organizationId" validate:"required"`
}

type CreateAdminPayload struct {
	Name  string `json:"name" gorm:"not null"`
	Email string `json:"email" gorm:"uniqueIndex;not null"`
}
type ActivateAdminPayload struct {
	Password        string `json:"password" gorm:"not null"`
	Token           string `json:"token" gorm:"uniqueIndex;not null"`
	ConfirmPassword string `json:"confirmPassword" gorm:"not null"`
}
type ActivateUserPayload struct {
	Password        string `json:"password" gorm:"not null"`
	Token           string `json:"token" gorm:"uniqueIndex;not null"`
	ConfirmPassword string `json:"confirmPassword" gorm:"not null"`
}

type ResendVerificationPayload struct {
	Email string `json:"email" gorm:"uniqueIndex;not null"`
}

type CreateOrganizationPayload struct {
	Name        string `json:"name" gorm:"not null"`
	Email       string `json:"email" gorm:"uniqueIndex;not null"`
	Description string `json:"description" gorm:"not null"`
	Website     string `json:"website" gorm:"not null"`
}
