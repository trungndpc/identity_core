package service

import (
	"errors"
	"testing"

	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

func TestParseZaloGraphProfileAcceptsNumericSuccess(t *testing.T) {
	profile, err := parseZaloGraphProfile([]byte(`{
		"error": 0,
		"message": "Success",
		"id": "2269534183882639390",
		"name": "Test User",
		"picture": {"data": {"url": "https://example.com/avatar.jpg"}}
	}`))
	if err != nil {
		t.Fatalf("parseZaloGraphProfile: %v", err)
	}
	if profile.ID != "2269534183882639390" || profile.Name != "Test User" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func TestParseZaloGraphProfileRejectsNumericError(t *testing.T) {
	_, err := parseZaloGraphProfile([]byte(`{
		"error": -201,
		"message": "Invalid access token"
	}`))
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != apperror.ErrUnauthorized.HTTPStatus {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestParseZaloGraphProfileRejectsObjectError(t *testing.T) {
	_, err := parseZaloGraphProfile([]byte(`{
		"error": {"code": 190, "message": "Token expired"}
	}`))
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != apperror.ErrUnauthorized.HTTPStatus {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}
