package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mightyfzeus/rbac/cmd/helpers"
	"github.com/mightyfzeus/rbac/internal/dtos"
	"github.com/mightyfzeus/rbac/internal/models"
	"github.com/mightyfzeus/rbac/internal/store"
	"go.uber.org/zap"
)

func (app *application) AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var payload dtos.LoginPayload
	if err := app.DecodeAndValidate(w, r, &payload); err != nil {
		return
	}

	admin, err := app.store.Admin.LoginAdmin(ctx, payload.Email, payload.Password)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if admin.Status == helpers.StatusPending {
		app.badRequestResponse(w, r, errors.New("Account yet to be activated, activate thy account"))
		return
	}

	token, err := GenerateJWT(admin.ID, admin.Email, admin.Name, string(admin.Role))
	if err != nil {
		app.internalServerError(w, r, err)
		app.logger.Error("error generating jwt token", zap.Error(err))
		return
	}

	app.jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"admin": admin,
		"token": token,
	}, "Admin login successful")
}

func (app *application) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := GetUserFromContext(ctx)
	if err != nil {
		app.unauthorizedResponse(w, r, err)
		return
	}
	if !app.HasPermission(user.Role, PermUsersCreate) {
		app.unauthorizedResponse(w, r, errors.New("you do not have permission to create users"))
		return
	}

	var payload dtos.CreateUserPayload
	if err = app.DecodeAndValidate(w, r, &payload); err != nil {
		return
	}

	inviteToken, err := app.GenerateInviteToken()
	if err != nil {
		app.internalServerError(w, r, err)
		app.logger.Error("error generating invite token", zap.Error(err))

		return
	}
	hashedToken := HashToken(inviteToken)
	if err != nil {
		app.internalServerError(w, r, err)
		app.logger.Error("error hashing invite token", zap.Error(err))
		return
	}

	// Check if the organization exists
	org, err := app.store.Organization.GetOrganization(ctx, payload.OrganizationID)

	if err != nil {
		app.badRequestResponse(w, r, errors.New("organization does not exist"))
		return
	}

	if !app.isOrgAdminOrSuper(user, org) {
		app.unauthorizedResponse(w, r, errors.New("unauthorized to add users to this organization"))
		return
	}

	newUser := &models.User{
		ID:             uuid.New(),
		Name:           payload.Email,
		Email:          payload.Email,
		Role:           RoleUser,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Status:         helpers.StatusPending,
		OrganizationID: org.ID,
	}

	err = app.store.WithTx(ctx, func(tx store.TxStorage) error {
		if err = tx.User.AddUserToOrganization(ctx, newUser); err != nil {
			return err
		}

		if err = tx.UserInvite.CreateUserInvites(ctx, &models.UserInvites{
			ID:        uuid.New(),
			UserId:    newUser.ID,
			TokenHash: hashedToken,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		}); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		app.logger.Info("error creating user", zap.Error(err))
		app.internalServerError(w, r, err)
		return
	}

	// TODO: user invites

	app.sendAdminInviteAsync(user.Email, inviteToken)

	app.jsonResponse(w, http.StatusCreated, nil, "Invite sent successfully")

}
func (app *application) CreateAdminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload dtos.CreateAdminPayload
	if err := app.DecodeAndValidate(w, r, &payload); err != nil {
		return
	}

	user, err := GetUserFromContext(ctx)
	if err != nil {
		app.unauthorizedResponse(w, r, err)
		return
	}
	if !app.HasPermission(user.Role, PermAdminCreate) {
		app.unauthorizedResponse(w, r, errors.New("you do not have permission to create admin"))
		return
	}

	inviteToken, err := app.GenerateInviteToken()
	if err != nil {
		app.internalServerError(w, r, err)
		app.logger.Error("error generating invite token", zap.Error(err))

		return
	}
	hashedToken := HashToken(inviteToken)
	if err != nil {
		app.internalServerError(w, r, err)
		app.logger.Error("error hashing invite token", zap.Error(err))
		return
	}

	// use transaction; it enables that ensures operation fails or succeed together so there are no orphaned records

	admin := &models.Admin{
		ID:    uuid.New(),
		Name:  payload.Email,
		Email: payload.Email,

		Role:      RoleAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    helpers.StatusPending,
		CreatedBy: uuid.MustParse(user.UserID),
	}

	err = app.store.WithTx(ctx, func(tx store.TxStorage) error {
		if err = tx.Admin.CreateAdmin(ctx, admin); err != nil {
			return err
		}
		invite := &models.AdminInvites{
			ID:        uuid.New(),
			AdminId:   admin.ID,
			TokenHash: hashedToken,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		}
		if err = tx.AdminInvites.CreateAdminInvites(ctx, invite); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		app.logger.Info("error creating admin", zap.Error(err))
		app.internalServerError(w, r, err)
		return
	}

	app.sendAdminInviteAsync(admin.Email, inviteToken)

	app.jsonResponse(w, http.StatusCreated, nil, "Invite resent successfully")

}

func (app *application) ActivateAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload dtos.ActivateAdminPayload
	if err := app.DecodeAndValidate(w, r, &payload); err != nil {
		return
	}
	if payload.Password != payload.ConfirmPassword {
		app.badRequestResponse(w, r, errors.New("passwords do not match"))
		return
	}
	// tokenHash := HashToken(payload.Token)

	invite, err := app.store.AdminInvites.ValidateToken(ctx, payload.Token)
	if err != nil || !invite.UsedAt.IsZero() || time.Now().After(invite.ExpiresAt) {
		app.badRequestResponse(w, r, errors.New("invalid or expired invite"))
		return
	}

	hashedPassword, err := HashPassword(payload.Password)
	if err != nil {
		app.logger.Error("error hashing password", zap.Error(err))
		app.internalServerError(w, r, err)
		return
	}

	err = app.store.WithTx(ctx, func(tx store.TxStorage) error {
		if err = tx.Admin.UpdateAdmin(ctx, invite.AdminId, map[string]interface{}{
			"password": hashedPassword,
			"status":   helpers.StatusActive,
		}); err != nil {
			return err
		}

		if err = tx.AdminInvites.UpdateInvite(ctx, invite.ID, map[string]interface{}{
			"used_at": time.Now(),
		}); err != nil {
			return err
		}

		return err
	})

	if err != nil {
		app.logger.Error("error activating admin", zap.Error(err))
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, nil, "Invite resent successfully")

}

func (app *application) ActivateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload dtos.ActivateUserPayload
	if err := app.DecodeAndValidate(w, r, &payload); err != nil {
		return
	}
	if payload.Password != payload.ConfirmPassword {
		app.badRequestResponse(w, r, errors.New("passwords do not match"))
		return
	}
	// tokenHash := HashToken(payload.Token)

	invite, err := app.store.UserInvite.ValidateUserToken(ctx, payload.Token)
	if err != nil || !invite.UsedAt.IsZero() || time.Now().After(invite.ExpiresAt) {
		app.badRequestResponse(w, r, errors.New("invalid or expired invite"))
		return
	}

	hashedPassword, err := HashPassword(payload.Password)
	if err != nil {
		app.logger.Error("error hashing password", zap.Error(err))
		app.internalServerError(w, r, err)
		return
	}

	err = app.store.WithTx(ctx, func(tx store.TxStorage) error {
		if err = tx.User.UpdateUser(ctx, invite.ID, map[string]interface{}{
			"password": hashedPassword,
			"status":   helpers.StatusActive,
		}); err != nil {
			return err
		}

		if err = tx.UserInvite.UpdateUserInvite(ctx, invite.ID, map[string]interface{}{
			"used_at": time.Now(),
		}); err != nil {
			return err
		}

		return err
	})

	if err != nil {
		app.logger.Error("error activating user", zap.Error(err))
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, nil, "User has been activated successfully")

}

func (app *application) ResendUserVerificationTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload dtos.ResendVerificationPayload
	if err := app.DecodeAndValidate(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	admin, err := app.store.Admin.GetAdminByEmail(ctx, payload.Email)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	rawToken, err := app.GenerateInviteToken()
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	tokenHash := HashToken(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	err = app.store.WithTx(ctx, func(tx store.TxStorage) error {
		invite, err2 := tx.AdminInvites.GetInviteByAdminId(ctx, admin.ID)
		if err2 != nil {
			return err2
		}
		if err != nil {
			return err
		}

		return tx.AdminInvites.UpdateInvite(ctx, invite.ID, map[string]interface{}{
			"token_hash": tokenHash,
			"expires_at": expiresAt,
			"used_at":    nil,
		})
	})

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.sendAdminInviteAsync(admin.Email, rawToken)

	app.jsonResponse(w, http.StatusOK, nil, "Invite resent successfully")
}

func (app *application) CreateOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload dtos.CreateOrganizationPayload
	if err := app.DecodeAndValidate(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := GetUserFromContext(ctx)
	if err != nil {
		app.unauthorizedResponse(w, r, err)
		return
	}

	if !app.HasPermission(user.Role, PermOrgCreate) {
		app.unauthorizedResponse(w, r, errors.New("unauthorized to create organization"))
		return
	}

	org := &models.Organization{
		ID:          uuid.New(),
		Name:        payload.Name,
		Email:       payload.Email,
		Description: payload.Description,
		Website:     payload.Website,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AdminID:     uuid.MustParse(user.UserID),
	}

	if err := app.store.Organization.CreateOrganization(ctx, org); err != nil {
		app.logger.Error("error creating organization", zap.Error(err))
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, org, "Organization created successfully")

}

func (app *application) GetOrganizationHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	user, err := GetUserFromContext(ctx)
	if err != nil {
		app.unauthorizedResponse(w, r, err)
		return
	}
	if !app.HasPermission(user.Role, PermOrgView) {
		app.unauthorizedResponse(w, r, errors.New("unauthorized to view organization"))
		return
	}

	orgId := r.URL.Query().Get("id")
	if orgId == "" {
		app.badRequestResponse(w, r, errors.New("Id is required"))
		return
	}
	parsedId, err := uuid.Parse(orgId)
	if err != nil {
		app.internalServerError(w, r, err)
		app.logger.Error("unable to parse id in delete org")
		return
	}

	org, err := app.store.Organization.GetOrganization(ctx, parsedId)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if !app.isOrgAdminOrSuper(user, org) {
		app.unauthorizedResponse(w, r, errors.New("unauthorized to access this organization"))
		return
	}

	app.jsonResponse(w, http.StatusOK, org, "organization")

}

func (app *application) DeleteOrganizationHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	user, err := GetUserFromContext(ctx)
	if err != nil {
		app.unauthorizedResponse(w, r, err)
		return
	}
	if !app.HasPermission(user.Role, PermOrgDelete) {
		app.unauthorizedResponse(w, r, errors.New("unauthorized to delete organization"))
		return
	}

	orgId := r.URL.Query().Get("id")
	if orgId == "" {
		app.badRequestResponse(w, r, errors.New("Id is required"))
		return
	}
	parsedId, err := uuid.Parse(orgId)
	if err != nil {
		app.internalServerError(w, r, err)
		app.logger.Error("unable to parse id in delete org")
		return
	}

	org, err := app.store.Organization.GetOrganization(ctx, parsedId)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if !app.isOrgAdminOrSuper(user, org) {
		app.unauthorizedResponse(w, r, errors.New("unauthorized to delete this organization"))
		return
	}

	err = app.store.Organization.DeleteOrganization(ctx, parsedId)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, nil, "organization deleted successfully")

}
