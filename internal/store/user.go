package store

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mightyfzeus/rbac/internal/models"
	"gorm.io/gorm"
)

type UserStore struct {
	db *gorm.DB
}

func (u *UserStore) AddUserToOrganization(ctx context.Context, user *models.User) error {
	err := u.db.WithContext(ctx).Create(user).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "Duplicate entry") {
			return ErrDuplicateEmail
		}
	}
	return err

}

func (a *UserStore) UpdateUser(
	ctx context.Context,
	userId uuid.UUID,
	updates map[string]interface{},
) error {
	return a.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userId).
		Updates(updates).
		Error
}

func (u *UserStore) LoginUser(ctx context.Context, email, password string) (*models.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		switch {
		case err.Error() == "record not found":
			return nil, ErrUserNotFound
		default:
			return nil, err
		}

	}

	if !CheckPassword(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil

}
