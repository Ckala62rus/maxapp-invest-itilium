// Package auth_test проверяет пакет internal/auth как внешний пользователь.
// Суффикс _test в имени пакета означает: тесты видят только экспортированные функции,
// поэтому мы проверяем публичный контракт, а не приватную реализацию.
package auth_test

import (
	// crypto/hmac и crypto/sha256 нужны, чтобы вручную собрать подпись initData
	// так же, как её должна прислать платформа MAX.
	"crypto/hmac"
	"crypto/sha256"
	// encoding/hex превращает бинарную HMAC-подпись в строку вида "a1b2...".
	"encoding/hex"
	// net/url используется для URL-encoding параметров initData.
	"net/url"
	// testing — стандартный пакет Go для тестов.
	"testing"
	// time нужен, чтобы тест не зависел от реальных часов компьютера.
	"time"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/auth"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/config"
	// require из testify останавливает тест сразу, если проверка не прошла.
	"github.com/stretchr/testify/require"
)

func TestManagerValidateInitData(t *testing.T) {
	// t.Parallel разрешает Go запускать этот тест параллельно с другими.
	// Так быстрее, но тест не должен менять глобальное состояние.
	t.Parallel()

	// Создаём менеджер авторизации с тестовым bot token.
	// Здесь не нужен весь config приложения — только AuthConfig.
	manager := auth.NewManager(config.AuthConfig{
		BotToken:       "test-bot-token",
		MaxInitDataTTL: 10 * time.Minute,
	})
	// Фиксируем "текущее" время, чтобы проверка TTL была повторяемой.
	now := time.Unix(1_800_000_000, 0).UTC()

	// signedInitData — helper ниже в этом файле. Он собирает строку initData
	// с корректной HMAC-подписью, как будто её прислал MAX.
	initData := signedInitData(t, "test-bot-token", map[string]string{
		"auth_date": "1800000000",
		"query_id":  "query-42",
		"user":      `{"id":12345,"first_name":"Max","last_name":"User","username":"max.user","language_code":"ru"}`,
	})

	// Act: вызываем проверяемый метод.
	result, err := manager.ValidateInitData(initData, now)

	// Assert: сначала проверяем, что ошибки нет.
	require.NoError(t, err)
	// Дальше проверяем, что менеджер правильно распарсил пользователя из initData.
	require.Equal(t, "12345", result.User.ID)
	require.Equal(t, "Max", result.User.FirstName)
	require.Equal(t, "User", result.User.LastName)
	require.Equal(t, "max.user", result.User.Username)
	require.Equal(t, "query-42", result.QueryID)
}

func TestManagerIssueAndParseAccessToken(t *testing.T) {
	t.Parallel()

	// В этом тесте проверяем пару операций:
	// 1. IssueAccessToken выпускает backend JWT-like токен.
	// 2. ParseAccessToken читает его обратно и возвращает те же claims.
	manager := auth.NewManager(config.AuthConfig{
		AccessTokenSecret: "super-secret",
		AccessTokenTTL:    2 * time.Hour,
	})
	now := time.Unix(1_800_000_000, 0).UTC()

	// Выпускаем токен для пользователя 100245.
	token, issuedClaims, err := manager.IssueAccessToken("100245", now)

	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.Equal(t, "100245", issuedClaims.UserID)

	// Парсим токен спустя час. TTL = 2 часа, значит токен ещё валиден.
	parsedClaims, err := manager.ParseAccessToken(token, now.Add(time.Hour))

	require.NoError(t, err)
	// Проверяем, что данные после парсинга совпали с тем, что было выпущено.
	require.Equal(t, issuedClaims.UserID, parsedClaims.UserID)
	require.Equal(t, issuedClaims.ExpiresAt, parsedClaims.ExpiresAt)
}

func signedInitData(t *testing.T, botToken string, params map[string]string) string {
	// t.Helper() говорит testing-пакету: если helper упадёт,
	// показывай ошибку на строке вызова helper-а, а не внутри helper-а.
	t.Helper()

	// encodedPairs пойдут в итоговую URL-encoded строку initData.
	encodedPairs := make([]string, 0, len(params))
	// launchPairs участвуют в data-check-string для подписи.
	// Важно: здесь значения НЕ URL-encoded.
	launchPairs := make([]string, 0, len(params))
	// Порядок ключей фиксирован, чтобы подпись была детерминированной.
	for _, key := range []string{"auth_date", "query_id", "user"} {
		value := params[key]
		encodedPairs = append(encodedPairs, key+"="+url.QueryEscape(value))
		launchPairs = append(launchPairs, key+"="+value)
	}

	// Первый HMAC получает secret key из bot token.
	secretMac := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secretMac.Write([]byte(botToken))
	secretKey := secretMac.Sum(nil)

	// Второй HMAC подписывает data-check-string.
	signatureMac := hmac.New(sha256.New, secretKey)
	_, _ = signatureMac.Write([]byte(launchPairs[0] + "\n" + launchPairs[1] + "\n" + launchPairs[2]))
	hash := hex.EncodeToString(signatureMac.Sum(nil))

	// MAX кладёт подпись в параметр hash.
	encodedPairs = append(encodedPairs, "hash="+url.QueryEscape(hash))

	// Возвращаем обычную query-string: key=value&key=value&hash=...
	return encodedPairs[0] + "&" + encodedPairs[1] + "&" + encodedPairs[2] + "&" + encodedPairs[3]
}
