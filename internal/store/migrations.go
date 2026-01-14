package store

import (
	"github.com/mightyfzeus/rbac/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Admin{},
		&models.Organization{},
		&models.AdminInvites{},
		&models.UserInvites{},
	)
}
