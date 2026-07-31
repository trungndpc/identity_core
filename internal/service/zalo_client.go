package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

type ZaloProfile struct {
	ID     string
	Name   string
	Avatar string
	Phone  string
}

type ZaloClient interface {
	GetProfile(ctx context.Context, accessToken string) (*ZaloProfile, error)
	ResolvePhoneNumber(ctx context.Context, accessToken, phoneToken string) (string, error)
}

type httpZaloClient struct {
	httpClient *http.Client
	secretKey  string
}

func NewHTTPZaloClient(secretKey string) ZaloClient {
	return &httpZaloClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		secretKey:  strings.TrimSpace(secretKey),
	}
}

type zaloGraphResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
	Error   json.RawMessage `json:"error"`
	Message string          `json:"message"`
}

type zaloGraphError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type zaloPhoneInfoResponse struct {
	Data *struct {
		Number string `json:"number"`
	} `json:"data"`
	Error   int    `json:"error"`
	Message string `json:"message"`
}

func (c *httpZaloClient) GetProfile(ctx context.Context, accessToken string) (*ZaloProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.zalo.me/v2.0/me?fields=id,name,picture", nil)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal.Code, "failed to build zalo request", apperror.ErrInternal.HTTPStatus)
	}
	req.Header.Set("access_token", accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal.Code, "failed to call zalo graph", apperror.ErrInternal.HTTPStatus)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal.Code, "failed to read zalo response", apperror.ErrInternal.HTTPStatus)
	}

	return parseZaloGraphProfile(body)
}

func parseZaloGraphProfile(body []byte) (*ZaloProfile, error) {
	var parsed zaloGraphResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal.Code, "invalid zalo response", apperror.ErrInternal.HTTPStatus)
	}

	if len(parsed.Error) > 0 && string(parsed.Error) != "null" {
		var code int
		if err := json.Unmarshal(parsed.Error, &code); err == nil {
			if code != 0 {
				message := strings.TrimSpace(parsed.Message)
				if message == "" {
					message = "zalo authentication failed"
				}
				return nil, apperror.New("ZALO_AUTH_FAILED", fmt.Sprintf("zalo error: %s", message), apperror.ErrUnauthorized.HTTPStatus)
			}
		} else {
			var graphError zaloGraphError
			if err := json.Unmarshal(parsed.Error, &graphError); err != nil {
				return nil, apperror.Wrap(err, apperror.ErrInternal.Code, "invalid zalo response", apperror.ErrInternal.HTTPStatus)
			}
			if graphError.Code != 0 || strings.TrimSpace(graphError.Message) != "" {
				message := strings.TrimSpace(graphError.Message)
				if message == "" {
					message = "zalo authentication failed"
				}
				return nil, apperror.New("ZALO_AUTH_FAILED", fmt.Sprintf("zalo error: %s", message), apperror.ErrUnauthorized.HTTPStatus)
			}
		}
	}

	if parsed.ID == "" {
		return nil, apperror.New("ZALO_AUTH_FAILED", "zalo profile missing id", apperror.ErrUnauthorized.HTTPStatus)
	}

	return &ZaloProfile{
		ID:     parsed.ID,
		Name:   parsed.Name,
		Avatar: parsed.Picture.Data.URL,
	}, nil
}

// ResolvePhoneNumber exchanges getPhoneNumber token via Zalo Graph.
// GET https://graph.zalo.me/v2.0/me/info
// Headers: access_token, code (phone token), secret_key (Zalo App secret)
func (c *httpZaloClient) ResolvePhoneNumber(ctx context.Context, accessToken, phoneToken string) (string, error) {
	if c.secretKey == "" {
		return "", apperror.New("ZALO_SECRET_MISSING", "ZALO_APP_SECRET_KEY is not configured", apperror.ErrInternal.HTTPStatus)
	}
	accessToken = strings.TrimSpace(accessToken)
	phoneToken = strings.TrimSpace(phoneToken)
	if accessToken == "" || phoneToken == "" {
		return "", apperror.New("ZALO_PHONE_TOKEN_REQUIRED", "access_token and phone_token are required", apperror.ErrBadRequest.HTTPStatus)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.zalo.me/v2.0/me/info", nil)
	if err != nil {
		return "", apperror.Wrap(err, apperror.ErrInternal.Code, "failed to build zalo phone request", apperror.ErrInternal.HTTPStatus)
	}
	req.Header.Set("access_token", accessToken)
	req.Header.Set("code", phoneToken)
	req.Header.Set("secret_key", c.secretKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", apperror.Wrap(err, apperror.ErrInternal.Code, "failed to call zalo phone api", apperror.ErrInternal.HTTPStatus)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperror.Wrap(err, apperror.ErrInternal.Code, "failed to read zalo phone response", apperror.ErrInternal.HTTPStatus)
	}

	var parsed zaloPhoneInfoResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", apperror.Wrap(err, apperror.ErrInternal.Code, "invalid zalo phone response", apperror.ErrInternal.HTTPStatus)
	}
	if parsed.Error != 0 {
		msg := parsed.Message
		if msg == "" {
			msg = "failed to resolve phone number"
		}
		return "", apperror.New("ZALO_PHONE_FAILED", msg, apperror.ErrBadRequest.HTTPStatus)
	}
	if parsed.Data == nil || strings.TrimSpace(parsed.Data.Number) == "" {
		return "", apperror.New("ZALO_PHONE_FAILED", "zalo phone number empty", apperror.ErrBadRequest.HTTPStatus)
	}

	return normalizeVNPhone(parsed.Data.Number), nil
}

func normalizeVNPhone(raw string) string {
	phone := strings.TrimSpace(raw)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	if strings.HasPrefix(phone, "84") && len(phone) >= 10 {
		return "0" + phone[2:]
	}
	if strings.HasPrefix(phone, "+84") && len(phone) >= 11 {
		return "0" + phone[3:]
	}
	return phone
}
