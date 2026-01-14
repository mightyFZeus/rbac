package store

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mightyfzeus/rbac/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CheckPassword(plain, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

type AdminStore struct {
	db *gorm.DB
}

func (u *AdminStore) LoginAdmin(ctx context.Context, email, password string) (*models.Admin, error) {
	var admin models.Admin
	err := u.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error
	if err != nil {
		switch {
		case err.Error() == "record not found":
			return nil, ErrUserNotFound
		default:
			return nil, err
		}

	}

	if !CheckPassword(password, admin.Password) {
		return nil, errors.New("invalid credentials")
	}

	return &admin, nil

}

func (u *AdminStore) CreateAdmin(ctx context.Context, user *models.Admin) error {
	err := u.db.WithContext(ctx).Create(user).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "Duplicate entry") {
			return ErrDuplicateEmail
		}
	}
	return err
}

func (a *AdminStore) GetAdmin(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	var admin models.Admin
	err := a.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
	}
	return &admin, err
}

func (a *AdminStore) UpdateAdmin(
	ctx context.Context,
	adminID uuid.UUID,
	updates map[string]interface{},
) error {
	return a.db.WithContext(ctx).
		Model(&models.Admin{}).
		Where("id = ?", adminID).
		Updates(updates).
		Error
}

func (a *AdminStore) GetAdminByEmail(ctx context.Context, email string) (*models.Admin, error) {
	var admin models.Admin
	err := a.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
	}
	return &admin, err
}
