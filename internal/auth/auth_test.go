package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"testing"
	"time"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/config"
	"github.com/stretchr/testify/require"
)

func TestManagerValidateInitData(t *testing.T) {
	t.Parallel()

	manager := NewManager(config.AuthConfig{
		BotToken:       "test-bot-token",
		MaxInitDataTTL: 10 * time.Minute,
	})
	now := time.Unix(1_800_000_000, 0).UTC()

	initData := signedInitData(t, "test-bot-token", map[string]string{
		"auth_date": "1800000000",
		"query_id":  "query-42",
		"user":      `{"id":12345,"first_name":"Max","last_name":"User","username":"max.user","language_code":"ru"}`,
	})

	result, err := manager.ValidateInitData(initData, now)

	require.NoError(t, err)
	require.Equal(t, "12345", result.User.ID)
	require.Equal(t, "Max", result.User.FirstName)
	require.Equal(t, "User", result.User.LastName)
	require.Equal(t, "max.user", result.User.Username)
	require.Equal(t, "query-42", result.QueryID)
}

func TestManagerIssueAndParseAccessToken(t *testing.T) {
	t.Parallel()

	manager := NewManager(config.AuthConfig{
		AccessTokenSecret: "super-secret",
		AccessTokenTTL:    2 * time.Hour,
	})
	now := time.Unix(1_800_000_000, 0).UTC()

	token, issuedClaims, err := manager.IssueAccessToken("100245", now)

	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.Equal(t, "100245", issuedClaims.UserID)

	parsedClaims, err := manager.ParseAccessToken(token, now.Add(time.Hour))

	require.NoError(t, err)
	require.Equal(t, issuedClaims.UserID, parsedClaims.UserID)
	require.Equal(t, issuedClaims.ExpiresAt, parsedClaims.ExpiresAt)
}

func signedInitData(t *testing.T, botToken string, params map[string]string) string {
	t.Helper()

	encodedPairs := make([]string, 0, len(params))
	launchPairs := make([]string, 0, len(params))
	for _, key := range []string{"auth_date", "query_id", "user"} {
		value := params[key]
		encodedPairs = append(encodedPairs, key+"="+url.QueryEscape(value))
		launchPairs = append(launchPairs, key+"="+value)
	}

	secretMac := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secretMac.Write([]byte(botToken))
	secretKey := secretMac.Sum(nil)

	signatureMac := hmac.New(sha256.New, secretKey)
	_, _ = signatureMac.Write([]byte(launchPairs[0] + "\n" + launchPairs[1] + "\n" + launchPairs[2]))
	hash := hex.EncodeToString(signatureMac.Sum(nil))

	encodedPairs = append(encodedPairs, "hash="+url.QueryEscape(hash))

	return encodedPairs[0] + "&" + encodedPairs[1] + "&" + encodedPairs[2] + "&" + encodedPairs[3]
}
