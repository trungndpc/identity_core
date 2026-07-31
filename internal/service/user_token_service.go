package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/updev/galaxy/identity_core/internal/config"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

const userTokenType = "user"

type UserPrincipal struct {
	UserID     int64
	TenantID   int64
	TenantCode string
}

type UserTokenService interface {
	IssueAccessToken(userID, tenantID int64, tenantCode string) (*dto.UserTokenResponse, error)
	ValidateAccessToken(tokenString string) (UserPrincipal, error)
}

type userTokenService struct {
	cfg *config.Config
}

type userJWTClaims struct {
	TenantID   int64  `json:"tenant_id"`
	TenantCode string `json:"tenant_code"`
	Typ        string `json:"typ"`
	jwt.RegisteredClaims
}

func NewUserTokenService(cfg *config.Config) UserTokenService {
	return &userTokenService{cfg: cfg}
}

func (s *userTokenService) IssueAccessToken(userID, tenantID int64, tenantCode string) (*dto.UserTokenResponse, error) {
	if s.cfg.UserJWTSecret == "" {
		return nil, apperror.ErrInternal
	}
	tenantCode = strings.TrimSpace(tenantCode)
	if userID <= 0 || tenantID <= 0 || tenantCode == "" {
		return nil, apperror.ErrUnauthorized
	}

	expiresIn := int64(s.cfg.UserJWTExpiryHours) * 3600
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	expiresAt := time.Now().UTC().Add(time.Duration(expiresIn) * time.Second)

	claims := userJWTClaims{
		TenantID:   tenantID,
		TenantCode: tenantCode,
		Typ:        userTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.UserJWTSecret))
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal.Code, "failed to sign user token", apperror.ErrInternal.HTTPStatus)
	}

	return &dto.UserTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}, nil
}

func (s *userTokenService) ValidateAccessToken(tokenString string) (UserPrincipal, error) {
	if s.cfg.UserJWTSecret == "" || tokenString == "" {
		return UserPrincipal{}, apperror.ErrUnauthorized
	}

	claims := &userJWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.UserJWTSecret), nil
	})
	if err != nil || !token.Valid || claims.Typ != userTokenType || claims.TenantCode == "" {
		return UserPrincipal{}, apperror.ErrUnauthorized
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 || claims.TenantID <= 0 {
		return UserPrincipal{}, apperror.ErrUnauthorized
	}

	return UserPrincipal{
		UserID:     userID,
		TenantID:   claims.TenantID,
		TenantCode: claims.TenantCode,
	}, nil
}
