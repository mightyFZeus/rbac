package store

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mightyfzeus/rbac/internal/models"
	"gorm.io/gorm"
)

type OrganizationStore struct {
	db *gorm.DB
}

func (o *OrganizationStore) CreateOrganization(ctx context.Context, org *models.Organization) error {
	err := o.db.WithContext(ctx).Create(org).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "Duplicate entry") {
			return ErrDuplicateOrgEmail
		}
	}

	return nil

}

func (o *OrganizationStore) GetOrganization(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	var organization models.Organization
	err := o.db.WithContext(ctx).Where("id = ?", id).First(&organization).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
	}
	return &organization, err
}

func (o *OrganizationStore) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	result := o.db.Where("id = ?", id).Delete(&models.Organization{})
	if result.Error != nil {
		o.db.Rollback()
		return result.Error
	}
	if result.RowsAffected == 0 {
		o.db.Rollback()
		return ErrOrgNotFound
	}

	return o.db.Commit().Error
}
