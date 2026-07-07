package auth

type Role string

const (
	RolePlatformAdmin Role = "platform_admin"
	RoleTenantAdmin   Role = "tenant_admin"
	RoleCustomer      Role = "customer"
)

func (r Role) Valid() bool {
	switch r {
	case RolePlatformAdmin, RoleTenantAdmin, RoleCustomer:
		return true
	default:
		return false
	}
}

func CanWriteKM(role Role) bool {
	return role == RolePlatformAdmin || role == RoleTenantAdmin
}