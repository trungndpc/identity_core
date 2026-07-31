package service

import (
	"testing"

	"github.com/updev/galaxy/identity_core/internal/config"
)

func TestUserTokenServiceIssueAndValidate(t *testing.T) {
	cfg := &config.Config{
		UserJWTSecret:      "test-user-secret",
		UserJWTExpiryHours: 1,
	}
	svc := NewUserTokenService(cfg)

	tokens, err := svc.IssueAccessToken(42, 7, "sgs")
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}
	if tokens.AccessToken == "" || tokens.TokenType != "Bearer" {
		t.Fatalf("unexpected tokens: %+v", tokens)
	}
	if tokens.ExpiresIn != 3600 {
		t.Fatalf("expected expires_in 3600, got %d", tokens.ExpiresIn)
	}

	principal, err := svc.ValidateAccessToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken: %v", err)
	}
	if principal.UserID != 42 || principal.TenantID != 7 || principal.TenantCode != "sgs" {
		t.Fatalf("unexpected principal: %+v", principal)
	}

	if _, err := svc.ValidateAccessToken("not-a-jwt"); err == nil {
		t.Fatal("expected invalid token error")
	}
}

func TestUserTokenServiceRejectsEmptyToken(t *testing.T) {
	cfg := &config.Config{
		UserJWTSecret:      "test-user-secret",
		UserJWTExpiryHours: 1,
	}
	svc := NewUserTokenService(cfg)
	if _, err := svc.ValidateAccessToken(""); err == nil {
		t.Fatal("expected empty token to fail")
	}
}
