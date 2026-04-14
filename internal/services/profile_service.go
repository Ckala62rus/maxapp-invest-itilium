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

// EmployeeLookupClient describes outbound employee lookup calls used by ProfileService.
type EmployeeLookupClient interface {
	// FindEmployeeByIdentifier requests a raw employee payload from ITILIUM.
	FindEmployeeByIdentifier(ctx context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error)
}

// ProfileService handles current user profile and registration flows.
type ProfileService struct {
	repository   ProfileRepository
	lookupClient EmployeeLookupClient
}

// NewProfileService creates a profile service.
func NewProfileService(repository ProfileRepository, lookupClient EmployeeLookupClient) *ProfileService {
	return &ProfileService{
		repository:   repository,
		lookupClient: lookupClient,
	}
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

// FindEmployeeByIdentifier loads a raw employee payload from ITILIUM for exploration and future identity mapping.
func (s *ProfileService) FindEmployeeByIdentifier(ctx context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error) {
	if strings.TrimSpace(request.Identifier) == "" {
		return models.EmployeeLookupResult{}, errors.New("identifier is required")
	}
	if strings.TrimSpace(request.AttributeCode) == "" {
		request.AttributeCode = "employee"
	}
	if s.lookupClient == nil {
		return models.EmployeeLookupResult{}, errors.New("employee lookup client is not configured")
	}

	return s.lookupClient.FindEmployeeByIdentifier(ctx, request)
}
