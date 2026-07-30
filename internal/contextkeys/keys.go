package contextkeys

type ContextKey string

const (
	RequestSource   ContextKey = "request_source"
	TenantID        ContextKey = "tenant_id"
	TenantCode      ContextKey = "tenant_code"
	UserID          ContextKey = "user_id"
	AdminUsername   ContextKey = "admin_username"
	AdminAPIKeyAuth ContextKey = "admin_api_key_auth"
)

const (
	SourceAdmin    = "admin"
	SourceInternal = "internal"
	SourceUser     = "user"
)
