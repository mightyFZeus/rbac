package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mightyfzeus/rbac/internal/models"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user does not exist")
	ErrDuplicateEmail    = errors.New("user with email already exists")
	ErrDuplicateOrgEmail = errors.New("organization with email already exists")
	ErrInvalidToken      = errors.New("invalid token ")
	ErrInviteNotFound    = errors.New("User does not have an invite ")
	ErrOrgNotFound       = errors.New("Organization not found")
)

type AdminStoreInterface interface {
	LoginAdmin(ctx context.Context, email, password string) (*models.Admin, error)
	CreateAdmin(ctx context.Context, user *models.Admin) error
	GetAdmin(ctx context.Context, id uuid.UUID) (*models.Admin, error)
	GetAdminByEmail(ctx context.Context, email string) (*models.Admin, error)
	UpdateAdmin(
		ctx context.Context,
		adminID uuid.UUID,
		updates map[string]interface{},
	) error
}

type AdminInviteStoreInterface interface {
	CreateAdminInvites(ctx context.Context, invite *models.AdminInvites) error
	ValidateToken(ctx context.Context, token string) (*models.AdminInvites, error)
	UpdateInvite(
		ctx context.Context,
		id uuid.UUID,
		updates map[string]interface{},
	) error
	GetInviteByAdminId(ctx context.Context, adminID uuid.UUID) (*models.AdminInvites, error)
}

type OrganizationStoreInterface interface {
	CreateOrganization(ctx context.Context, org *models.Organization) error
	GetOrganization(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
}

type UserStoreInterface interface {
	AddUserToOrganization(ctx context.Context, user *models.User) error
	UpdateUser(
		ctx context.Context,
		userId uuid.UUID,
		updates map[string]interface{},
	) error
	LoginUser(ctx context.Context, email, password string) (*models.User, error)
}

type UserInviteStoreInterface interface {
	CreateUserInvites(ctx context.Context, invite *models.UserInvites) error
	ValidateUserToken(ctx context.Context, token string) (*models.UserInvites, error)
	UpdateUserInvite(
		ctx context.Context,
		id uuid.UUID,
		updates map[string]interface{},
	) error
	GetInviteByUserId(ctx context.Context, userId uuid.UUID) (*models.UserInvites, error)
}

type Storage struct {
	Admin        AdminStoreInterface
	AdminInvites AdminInviteStoreInterface
	Organization OrganizationStoreInterface
	User         UserStoreInterface
	UserInvite   UserInviteStoreInterface
}

func NewStorage(db *gorm.DB) Storage {
	return Storage{
		Admin:        &AdminStore{db: db},
		AdminInvites: &AdminInviteStore{db: db},
		Organization: &OrganizationStore{db: db},
		User:         &UserStore{db: db},
		UserInvite:   &UserInviteStore{db: db},
	}
}

type TxStorage struct {
	Admin        AdminStoreInterface
	AdminInvites AdminInviteStoreInterface
	Organization OrganizationStoreInterface
	User         UserStoreInterface
	UserInvite   UserInviteStoreInterface
}

func (s Storage) WithTx(ctx context.Context, fn func(tx TxStorage) error) error {
	tx := s.Admin.(*AdminStore).db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txs := TxStorage{
		Admin:        &AdminStore{db: tx},
		AdminInvites: &AdminInviteStore{db: tx},
		Organization: &OrganizationStore{db: tx},
		User:         &UserStore{db: tx},
		UserInvite:   &UserInviteStore{db: tx},
	}

	err := fn(txs)
	if err != nil {
		rbErr := tx.Rollback().Error
		if rbErr != nil {
			return errors.New(err.Error() + " | rollback error: " + rbErr.Error())
		}
		return err
	}

	return tx.Commit().Error
}
