package store

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mightyfzeus/rbac/internal/models"
	"gorm.io/gorm"
)

type AdminInviteStore struct {
	db *gorm.DB
}

func (a *AdminInviteStore) CreateAdminInvites(ctx context.Context, invite *models.AdminInvites) error {
	err := a.db.WithContext(ctx).Create(invite).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "Duplicate entry") {
			return ErrDuplicateEmail
		}
	}
	return err
}

func (a *AdminInviteStore) ValidateToken(ctx context.Context, token string) (*models.AdminInvites, error) {
	var invite models.AdminInvites
	err := a.db.WithContext(ctx).Where("token_hash = ?", token).First(&invite).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
	}
	return &invite, err
}

func (a *AdminInviteStore) UpdateInvite(
	ctx context.Context,
	id uuid.UUID,
	updates map[string]interface{},
) error {
	return a.db.WithContext(ctx).
		Model(&models.AdminInvites{}).
		Where("id = ?", id).
		Updates(updates).
		Error
}

func (a *AdminInviteStore) GetInviteByAdminId(ctx context.Context, adminID uuid.UUID) (*models.AdminInvites, error) {
	var invite models.AdminInvites
	err := a.db.WithContext(ctx).Where("admin_id = ?", adminID).First(&invite).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInviteNotFound
		}
	}
	return &invite, err
}
