package main

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
)

const (
	PermUsersCreate    = "users:create"
	PermUsersUpdate    = "users:update"
	PermAdminCreate    = "admin:create"
	PermAdminUpdate    = "admin:update"
	PermAdminDelete    = "admin:delete"
	PermUsersDelete    = "users:delete"
	PermRolesAssign    = "roles:assign"
	PermSettingsSystem = "settings:system"
	PermSettingsOrg    = "settings:org"
	PermLogsView       = "logs:view"

	PermPostsCreate = "posts:create"
	PermPostsUpdate = "posts:update"
	PermPostsDelete = "posts:delete"

	PermOrgCreate  = "organization:create"
	PermOrgView    = "organization:view"
	PermOrgUpdate  = "organization:update"
	PermOrgDelete  = "organization:delete"
	PermOrgSuspend = "organization:suspend"
)

var RolePermissions = map[string][]string{
	RoleSuperAdmin: {
		PermAdminCreate,
		PermAdminUpdate,
		PermAdminDelete,
		PermUsersDelete,
		PermRolesAssign,
		PermSettingsSystem,
		PermSettingsOrg,
		PermLogsView,
		PermPostsUpdate,
		PermPostsDelete,
	},
	RoleAdmin: {
		PermUsersCreate,
		PermUsersUpdate,
		PermUsersDelete,
		PermRolesAssign,
		PermSettingsOrg,
		PermLogsView,
		PermPostsUpdate,
		PermPostsDelete,

		PermOrgCreate,
		PermOrgView,
		PermOrgUpdate,
		PermOrgDelete,
		PermOrgSuspend,
	},
	RoleUser: {
		PermPostsCreate, PermPostsUpdate, PermPostsDelete,
	},
}
