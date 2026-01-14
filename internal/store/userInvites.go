package store

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mightyfzeus/rbac/internal/models"
	"gorm.io/gorm"
)

type UserInviteStore struct {
	db *gorm.DB
}

func (a *UserInviteStore) CreateUserInvites(ctx context.Context, invite *models.UserInvites) error {
	err := a.db.WithContext(ctx).Create(invite).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "Duplicate entry") {
			return ErrDuplicateEmail
		}
	}
	return err
}

func (a *UserInviteStore) ValidateUserToken(ctx context.Context, token string) (*models.UserInvites, error) {
	var invite models.UserInvites
	err := a.db.WithContext(ctx).Where("token_hash = ?", token).First(&invite).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
	}
	return &invite, err
}

func (a *UserInviteStore) UpdateUserInvite(
	ctx context.Context,
	id uuid.UUID,
	updates map[string]interface{},
) error {
	return a.db.WithContext(ctx).
		Model(&models.UserInvites{}).
		Where("id = ?", id).
		Updates(updates).
		Error
}

func (a *UserInviteStore) GetInviteByUserId(ctx context.Context, userId uuid.UUID) (*models.UserInvites, error) {
	var invite models.UserInvites
	err := a.db.WithContext(ctx).Where("user_id = ?", userId).First(&invite).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInviteNotFound
		}
	}
	return &invite, err
}
