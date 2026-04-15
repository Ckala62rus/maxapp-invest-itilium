// Package services contains business logic orchestrators.
package services

import (
	"context"
	"errors"
	"strings"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/api"
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
	// RegisterUser sends a registration request to ITILIUM.
	RegisterUser(ctx context.Context, request models.RegistrationRequest) error
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

	if profile, ok := s.repository.GetByUserID(ctx, userID); ok && ((profile.EmployeeFound && !profile.RegistrationRequired) || profile.RegistrationPending) {
		return profile, nil
	}

	if s.lookupClient != nil {
		lookup, err := s.FindEmployeeByIdentifier(ctx, models.EmployeeLookupRequest{
			Identifier: userID,
		})
		if err != nil {
			if profile, ok := profileFromLookupError(userID, err); ok {
				return profile, nil
			}
			return models.UserProfile{}, err
		}

		if profile, ok := profileFromLookup(userID, lookup); ok {
			return profile, nil
		}
	}

	return fallbackProfile(userID), nil
}

// Register stores a registration request and returns the newly linked profile snapshot.
func (s *ProfileService) Register(ctx context.Context, request models.RegistrationRequest) (models.UserProfile, error) {
	if strings.TrimSpace(request.UserID) == "" {
		return models.UserProfile{}, errors.New("user id is required")
	}
	if strings.TrimSpace(request.FullName) == "" {
		return models.UserProfile{}, errors.New("full name is required")
	}
	if strings.TrimSpace(request.Organization) == "" {
		return models.UserProfile{}, errors.New("organization is required")
	}
	if strings.TrimSpace(request.Department) == "" {
		return models.UserProfile{}, errors.New("department is required")
	}
	if strings.TrimSpace(request.Position) == "" {
		return models.UserProfile{}, errors.New("position is required")
	}
	if s.lookupClient == nil {
		return models.UserProfile{}, errors.New("employee lookup client is not configured")
	}

	if err := s.lookupClient.RegisterUser(ctx, request); err != nil {
		return models.UserProfile{}, err
	}

	return s.repository.SaveRegistration(ctx, request), nil
}

// FindEmployeeByIdentifier loads a raw employee payload from ITILIUM for exploration and future identity mapping.
func (s *ProfileService) FindEmployeeByIdentifier(ctx context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error) {
	if strings.TrimSpace(request.Identifier) == "" {
		return models.EmployeeLookupResult{}, errors.New("identifier is required")
	}
	if strings.TrimSpace(request.AttributeCode) == "" {
		request.AttributeCode = "id"
	}
	if s.lookupClient == nil {
		return models.EmployeeLookupResult{}, errors.New("employee lookup client is not configured")
	}

	return s.lookupClient.FindEmployeeByIdentifier(ctx, request)
}

// profileFromLookup converts a normalized employee lookup payload into the main profile response.
func profileFromLookup(userID string, lookup models.EmployeeLookupResult) (models.UserProfile, bool) {
	firstName := firstLookupValue(lookup.Raw, "firstName")
	lastName := firstLookupValue(lookup.Raw, "lastName")
	middleName := firstLookupValue(lookup.Raw, "middleName")
	fullName := strings.TrimSpace(strings.Join([]string{firstName, lastName, middleName}, " "))
	if fullName == "" {
		fullName = firstLookupValue(lookup.Raw, "displayName", "fullName", "name", "fio", "FIO")
	}
	username := firstLookupValue(lookup.Raw, "username", "login", "telegram", "id")
	department := firstLookupValue(lookup.Raw, "client", "department", "subdivision", "division", "store", "shop", "filial")
	organization := firstLookupValue(lookup.Raw, "OU", "organization", "Organization")
	position := firstLookupValue(lookup.Raw, "post", "position", "NamePosition")

	found := strings.TrimSpace(lookup.UUID) != "" || fullName != "" || department != "" || len(lookup.ServiceCalls) > 0
	if !found {
		return models.UserProfile{}, false
	}

	if username == "" {
		username = userID
	}
	if fullName == "" {
		fullName = username
	}
	if department == "" {
		department = "Не определено"
	}

	return models.UserProfile{
		UserID:                    userID,
		Username:                  username,
		FullName:                  fullName,
		FirstName:                 firstName,
		LastName:                  lastName,
		MiddleName:                middleName,
		Department:                department,
		Organization:              organization,
		Position:                  position,
		ServiceCalls:              lookup.ServiceCalls,
		CanCreateMarketingRequests: lookup.CanCreateMarketingRequests,
		CanCreateDaxRequests:      lookup.CanCreateDaxRequests,
		EmployeeFound:             true,
		RegistrationRequired:      false,
	}, true
}

// fallbackProfile keeps the mini app usable when ITILIUM does not know the current user yet.
func fallbackProfile(userID string) models.UserProfile {
	return models.UserProfile{
		UserID:               userID,
		Username:             userID,
		FullName:             "Неизвестный пользователь",
		Department:           "Не определено",
		EmployeeFound:        false,
		RegistrationRequired: true,
		StatusMessage:        "Пользователь не найден в ITILIUM. Требуется регистрация.",
	}
}

// profileFromLookupError converts known ITILIUM status codes into UI-friendly onboarding states.
func profileFromLookupError(userID string, err error) (models.UserProfile, bool) {
	var statusErr api.HTTPStatusError
	if !errors.As(err, &statusErr) {
		return models.UserProfile{}, false
	}

	switch statusErr.StatusCode {
	case 401, 404:
		return fallbackProfile(userID), true
	case 403:
		return models.UserProfile{
			UserID:               userID,
			Username:             userID,
			FullName:             "Заявка на регистрацию отправлена",
			Department:           "Не определено",
			EmployeeFound:        false,
			RegistrationRequired: false,
			RegistrationPending:  true,
			StatusMessage:        "Ваша заявка на регистрацию еще на рассмотрении.",
		}, true
	default:
		return models.UserProfile{}, false
	}
}

// firstLookupValue returns the first non-empty string-like field from a raw ITILIUM payload.
func firstLookupValue(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}

		text := strings.TrimSpace(stringFromLookupValue(value))
		if text != "" {
			return text
		}
	}

	return ""
}

// stringFromLookupValue converts dynamic ITILIUM fields into strings without panics.
func stringFromLookupValue(value any) string {
	switch converted := value.(type) {
	case string:
		return converted
	case nil:
		return ""
	default:
		return ""
	}
}
