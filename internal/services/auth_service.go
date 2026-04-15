package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/auth"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
)

// AuthService validates MAX initData and exchanges it for backend access tokens.
type AuthService struct {
	manager *auth.Manager
	logger  *slog.Logger
}

// NewAuthService creates a new auth service.
func NewAuthService(manager *auth.Manager, logger *slog.Logger) *AuthService {
	return &AuthService{
		manager: manager,
		logger:  logger,
	}
}

// ValidateMaxInitData checks MAX initData and returns a backend access token.
func (s *AuthService) ValidateMaxInitData(_ context.Context, initData string) (models.MaxAuthValidateResponse, error) {
	if s.manager == nil {
		return models.MaxAuthValidateResponse{}, errors.New("auth manager is not configured")
	}
	if strings.TrimSpace(initData) == "" {
		return models.MaxAuthValidateResponse{}, errors.New("max init data is required")
	}

	if s.logger != nil {
		s.logger.Info(
			"max auth validation started",
			"init_data_length", len(initData),
		)
	}

	now := time.Now().UTC()
	result, err := s.manager.ValidateInitData(initData, now)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn(
				"max auth validation failed",
				"init_data_length", len(initData),
				"error", err,
			)
		}
		return models.MaxAuthValidateResponse{}, err
	}

	accessToken, claims, err := s.manager.IssueAccessToken(result.User.ID, now)
	if err != nil {
		if s.logger != nil {
			s.logger.Error(
				"max auth token issue failed",
				"user_id", result.User.ID,
				"query_id", result.QueryID,
				"error", err,
			)
		}
		return models.MaxAuthValidateResponse{}, err
	}

	if s.logger != nil {
		s.logger.Info(
			"max auth validation succeeded",
			"user_id", result.User.ID,
			"query_id", result.QueryID,
			"auth_date", result.AuthDate,
			"expires_at", claims.ExpiresAt,
		)
		s.logger.Debug(
			"max auth validation details",
			"user_id", result.User.ID,
			"raw_fields", result.Raw,
		)
	}

	return models.MaxAuthValidateResponse{
		AccessToken: accessToken,
		ExpiresAt:   claims.ExpiresAt,
		Identity: models.MaxIdentity{
			UserID:    result.User.ID,
			Username:  strings.TrimSpace(result.User.Username),
			FullName:  result.User.FullName(),
			FirstName: strings.TrimSpace(result.User.FirstName),
			LastName:  strings.TrimSpace(result.User.LastName),
		},
	}, nil
}
