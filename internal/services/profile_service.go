// Package services contains business logic orchestrators.
package services

import (
	"context"
	"errors"
	"strings"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
)

// ProfileRepository describes the storage operations used by ProfileService.
type ProfileRepository interface {
	// GetByUserID returns a stored profile by MAX user id.
	GetByUserID(ctx context.Context, userID string) (models.UserProfile, bool)
	// SaveRegistration stores a submitted registration form and returns the new profile snapshot.
	SaveRegistration(ctx context.Context, request models.RegistrationRequest) models.UserProfile
}

// ProfileService handles current user profile and registration flows.
type ProfileService struct {
	repository ProfileRepository
}

// NewProfileService creates a profile service.
func NewProfileService(repository ProfileRepository) *ProfileService {
	return &ProfileService{repository: repository}
}

// GetProfile returns the profile snapshot for a MAX user.
func (s *ProfileService) GetProfile(ctx context.Context, userID string) (models.UserProfile, error) {
	if strings.TrimSpace(userID) == "" {
		return models.UserProfile{}, errors.New("user id is required")
	}

	if profile, ok := s.repository.GetByUserID(ctx, userID); ok {
		return profile, nil
	}

	return models.UserProfile{
		UserID:               userID,
		Username:             userID,
		FullName:             "Неизвестный пользователь",
		Department:           "Не определено",
		EmployeeFound:        false,
		RegistrationRequired: true,
	}, nil
}

// Register stores a registration request and returns the newly linked profile snapshot.
func (s *ProfileService) Register(ctx context.Context, request models.RegistrationRequest) (models.UserProfile, error) {
	if strings.TrimSpace(request.UserID) == "" {
		return models.UserProfile{}, errors.New("user id is required")
	}
	if strings.TrimSpace(request.FullName) == "" {
		return models.UserProfile{}, errors.New("full name is required")
	}
	if strings.TrimSpace(request.Department) == "" {
		return models.UserProfile{}, errors.New("department is required")
	}

	return s.repository.SaveRegistration(ctx, request), nil
}
