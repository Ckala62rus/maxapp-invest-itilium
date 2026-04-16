// Package auth validates MAX WebAppData and issues backend access tokens.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/config"
)

// MaxUser stores the normalized MAX user identity parsed from WebAppData.
type MaxUser struct {
	// ID stores the trusted MAX user id.
	ID string `json:"id"`
	// FirstName stores the user's first name when MAX provided it.
	FirstName string `json:"first_name"`
	// LastName stores the user's last name when MAX provided it.
	LastName string `json:"last_name"`
	// Username stores the MAX username or nickname when MAX provided it.
	Username string `json:"username"`
	// LanguageCode stores the user's UI locale when MAX provided it.
	LanguageCode string `json:"language_code"`
	// PhotoURL stores the user's avatar URL when MAX provided it.
	PhotoURL string `json:"photo_url"`
}

// FullName joins the available MAX name parts into one display string.
func (u MaxUser) FullName() string {
	return strings.TrimSpace(strings.TrimSpace(u.FirstName) + " " + strings.TrimSpace(u.LastName))
}

// ValidateResult stores the parsed trusted identity after MAX initData validation.
type ValidateResult struct {
	// User stores the normalized MAX user identity.
	User MaxUser
	// AuthDate stores the timestamp when MAX signed the launch payload.
	AuthDate time.Time
	// QueryID stores the platform request correlation identifier.
	QueryID string
	// Raw stores the decoded key/value payload for logging and diagnostics.
	Raw map[string]string
}

// AccessTokenClaims stores the backend session payload signed into the bearer token.
type AccessTokenClaims struct {
	// UserID stores the trusted MAX user id.
	UserID string `json:"userId"`
	// IssuedAt stores the unix timestamp when the token was created.
	IssuedAt int64 `json:"issuedAt"`
	// ExpiresAt stores the unix timestamp when the token expires.
	ExpiresAt int64 `json:"expiresAt"`
}

// Manager validates MAX initData and signs backend access tokens.
type Manager struct {
	cfg config.AuthConfig
}

// NewManager creates an auth manager from runtime configuration.
func NewManager(cfg config.AuthConfig) *Manager {
	return &Manager{cfg: cfg}
}

// ValidateInitData verifies MAX WebAppData and extracts a trusted user identity.
func (m *Manager) ValidateInitData(initData string, now time.Time) (ValidateResult, error) {
	if strings.TrimSpace(m.cfg.BotToken) == "" {
		return ValidateResult{}, errors.New("max bot token is not configured")
	}
	if strings.TrimSpace(initData) == "" {
		return ValidateResult{}, errors.New("max init data is required")
	}

	// Строка initData — набор key=value, отдельно переданный hash; парсим в map без поля hash.
	params, providedHash, err := parseInitData(initData)
	if err != nil {
		return ValidateResult{}, err
	}

	// Алгоритм как у Telegram WebApp: secret_key = HMAC_SHA256("WebAppData", bot_token), затем HMAC по отсортированным парам.
	computedHash := computeInitDataHash(params, m.cfg.BotToken)
	if !hmac.Equal([]byte(strings.ToLower(providedHash)), []byte(computedHash)) {
		return ValidateResult{}, errors.New("max init data signature is invalid")
	}

	authDateValue, ok := params["auth_date"]
	if !ok {
		return ValidateResult{}, errors.New("max init data auth_date is required")
	}

	authDateUnix, err := strconv.ParseInt(authDateValue, 10, 64)
	if err != nil {
		return ValidateResult{}, errors.New("max init data auth_date is invalid")
	}

	authDate := time.Unix(authDateUnix, 0).UTC()
	maxAge := m.cfg.MaxInitDataTTL
	if maxAge <= 0 {
		maxAge = 10 * time.Minute
	}
	// Небольшой допуск по часам, чтобы не отбраковывать из-за рассинхрона времени.
	if authDate.After(now.Add(1 * time.Minute)) {
		return ValidateResult{}, errors.New("max init data auth_date is in the future")
	}
	if now.Sub(authDate) > maxAge {
		return ValidateResult{}, errors.New("max init data has expired")
	}

	userValue, ok := params["user"]
	if !ok || strings.TrimSpace(userValue) == "" {
		return ValidateResult{}, errors.New("max init data user is required")
	}

	user, err := parseMaxUser(userValue)
	if err != nil {
		return ValidateResult{}, err
	}

	return ValidateResult{
		User:     user,
		AuthDate: authDate,
		QueryID:  params["query_id"],
		Raw:      params,
	}, nil
}

// IssueAccessToken signs a backend bearer token for the trusted MAX user identity.
func (m *Manager) IssueAccessToken(userID string, now time.Time) (string, AccessTokenClaims, error) {
	if strings.TrimSpace(m.cfg.AccessTokenSecret) == "" {
		return "", AccessTokenClaims{}, errors.New("auth access token secret is not configured")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", AccessTokenClaims{}, errors.New("user id is required")
	}

	ttl := m.cfg.AccessTokenTTL
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}

	// Формат токена: base64url(JSON claims) + "." + HMAC-SHA256(payload, secret), без сторонних библиотек JWT.
	claims := AccessTokenClaims{
		UserID:    userID,
		IssuedAt:  now.UTC().Unix(),
		ExpiresAt: now.UTC().Add(ttl).Unix(),
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", AccessTokenClaims{}, fmt.Errorf("encode auth token claims: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := signAccessToken(encodedPayload, m.cfg.AccessTokenSecret)

	return encodedPayload + "." + signature, claims, nil
}

// ParseAccessToken validates a backend bearer token and returns the trusted claims.
func (m *Manager) ParseAccessToken(token string, now time.Time) (AccessTokenClaims, error) {
	if strings.TrimSpace(m.cfg.AccessTokenSecret) == "" {
		return AccessTokenClaims{}, errors.New("auth access token secret is not configured")
	}

	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 {
		return AccessTokenClaims{}, errors.New("access token format is invalid")
	}

	// Сверяем подпись константным временем (hmac.Equal), затем проверяем срок действия.
	expectedSignature := signAccessToken(parts[0], m.cfg.AccessTokenSecret)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[1])) {
		return AccessTokenClaims{}, errors.New("access token signature is invalid")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return AccessTokenClaims{}, errors.New("access token payload is invalid")
	}

	var claims AccessTokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return AccessTokenClaims{}, errors.New("access token claims are invalid")
	}

	if strings.TrimSpace(claims.UserID) == "" {
		return AccessTokenClaims{}, errors.New("access token user id is required")
	}
	if claims.ExpiresAt <= now.UTC().Unix() {
		return AccessTokenClaims{}, errors.New("access token has expired")
	}

	return claims, nil
}

func parseInitData(initData string) (map[string]string, string, error) {
	pairs := strings.Split(initData, "&")
	params := make(map[string]string, len(pairs))
	var providedHash string

	// Ключ hash выделяем: он не участвует в расчёте подписи полезной нагрузки.
	for _, pair := range pairs {
		if strings.TrimSpace(pair) == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, "", errors.New("max init data pair is invalid")
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, "", errors.New("max init data key is invalid")
		}
		if _, exists := params[key]; exists {
			return nil, "", errors.New("max init data contains duplicate keys")
		}

		value, err := url.QueryUnescape(parts[1])
		if err != nil {
			return nil, "", fmt.Errorf("decode max init data value: %w", err)
		}

		if key == "hash" {
			providedHash = strings.ToLower(strings.TrimSpace(value))
			continue
		}

		params[key] = value
	}

	if providedHash == "" {
		return nil, "", errors.New("max init data hash is required")
	}

	return params, providedHash, nil
}

func computeInitDataHash(params map[string]string, botToken string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// data_check_string: ключи по алфавиту, строки "key=value" через \n.
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, key+"="+params[key])
	}

	launchParams := strings.Join(lines, "\n")
	secretKey := computeHMAC([]byte("WebAppData"), []byte(botToken))
	signature := computeHMAC(secretKey, []byte(launchParams))

	return hex.EncodeToString(signature)
}

func parseMaxUser(rawUser string) (MaxUser, error) {
	var payload struct {
		ID           any    `json:"id"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Username     string `json:"username"`
		LanguageCode string `json:"language_code"`
		PhotoURL     string `json:"photo_url"`
	}

	if err := json.Unmarshal([]byte(rawUser), &payload); err != nil {
		return MaxUser{}, errors.New("max init data user payload is invalid")
	}

	userID := stringifyID(payload.ID)
	if userID == "" {
		return MaxUser{}, errors.New("max init data user id is required")
	}

	return MaxUser{
		ID:           userID,
		FirstName:    strings.TrimSpace(payload.FirstName),
		LastName:     strings.TrimSpace(payload.LastName),
		Username:     strings.TrimSpace(payload.Username),
		LanguageCode: strings.TrimSpace(payload.LanguageCode),
		PhotoURL:     strings.TrimSpace(payload.PhotoURL),
	}, nil
}

func stringifyID(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	default:
		return ""
	}
}

func signAccessToken(payload string, secret string) string {
	return base64.RawURLEncoding.EncodeToString(computeHMAC([]byte(secret), []byte(payload)))
}

func computeHMAC(key []byte, payload []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}
