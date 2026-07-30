package domain

const (
	TenantStatusActive   = "active"
	TenantStatusInactive = "inactive"
	TenantStatusSuspended = "suspended"

	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"

	RelationshipStatusActive   = "active"
	RelationshipStatusInactive = "inactive"
	RelationshipStatusPending  = "pending"

	IdentityProviderLocal = "local"
	IdentityProviderGoogle = "google"
	IdentityProviderZalo   = "zalo"

	IdentityTypeEmail    = "email"
	IdentityTypePhone    = "phone"
	IdentityTypeUsername = "username"
	IdentityTypeZaloID   = "zalo_id"
)
