package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/api"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/services"
	"github.com/stretchr/testify/require"
)

type profileRepositoryStub struct {
	getProfile models.UserProfile
	getOK      bool
	saved      models.RegistrationRequest
}

type employeeLookupClientStub struct {
	response models.EmployeeLookupResult
	err      error
	regErr   error
}

func (s *profileRepositoryStub) GetByUserID(_ context.Context, _ string) (models.UserProfile, bool) {
	return s.getProfile, s.getOK
}

func (s *profileRepositoryStub) SaveRegistration(_ context.Context, request models.RegistrationRequest) models.UserProfile {
	s.saved = request
	return models.UserProfile{
		UserID:               request.UserID,
		Username:             request.UserID,
		FullName:             request.FullName,
		Department:           request.Department,
		EmployeeFound:        false,
		RegistrationRequired: false,
		RegistrationPending:  true,
		StatusMessage:        "Ваша заявка на регистрацию еще на рассмотрении.",
	}
}

func (s *employeeLookupClientStub) FindEmployeeByIdentifier(_ context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error) {
	if s.err != nil {
		return models.EmployeeLookupResult{}, s.err
	}
	if s.response.AttributeCode == "" {
		s.response.AttributeCode = request.AttributeCode
	}
	if s.response.Identifier == "" {
		s.response.Identifier = request.Identifier
	}

	return s.response, nil
}

func (s *employeeLookupClientStub) RegisterUser(_ context.Context, _ models.RegistrationRequest) error {
	return s.regErr
}

func TestProfileServiceGetProfileFallback(t *testing.T) {
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{})

	profile, err := service.GetProfile(context.Background(), "100500")

	require.NoError(t, err)
	require.Equal(t, "100500", profile.UserID)
	require.True(t, profile.RegistrationRequired)
}

func TestProfileServiceGetProfileUsesLookupWhenStoredProfileNeedsRegistration(t *testing.T) {
	service := services.NewProfileService(&profileRepositoryStub{
		getProfile: models.UserProfile{
			UserID:               "100245",
			Username:             "100245",
			FullName:             "Неизвестный пользователь",
			Department:           "Не определено",
			EmployeeFound:        false,
			RegistrationRequired: true,
		},
		getOK: true,
	}, &employeeLookupClientStub{
		response: models.EmployeeLookupResult{
			UUID:                       "emp-42",
			ServiceCalls:               []string{"0000017195", "0000019683"},
			CanCreateMarketingRequests: false,
			CanCreateDaxRequests:       true,
			Raw: map[string]any{
				"firstName":  "Овчинников",
				"lastName":   "Евгений",
				"middleName": "Аркадьевич",
				"post":       "Веб-программист",
				"OU":         "ООО \"ИНВЕСТ ИТ\"",
				"client":     "Отдел информационных технологий",
				"id":         "40367639",
			},
		},
	})

	profile, err := service.GetProfile(context.Background(), "100245")

	require.NoError(t, err)
	require.Equal(t, "100245", profile.UserID)
	require.Equal(t, "40367639", profile.Username)
	require.Equal(t, "Овчинников Евгений Аркадьевич", profile.FullName)
	require.Equal(t, "Овчинников", profile.FirstName)
	require.Equal(t, "Евгений", profile.LastName)
	require.Equal(t, "Аркадьевич", profile.MiddleName)
	require.Equal(t, "Отдел информационных технологий", profile.Department)
	require.Equal(t, "ООО \"ИНВЕСТ ИТ\"", profile.Organization)
	require.Equal(t, "Веб-программист", profile.Position)
	require.Len(t, profile.ServiceCalls, 2)
	require.False(t, profile.CanCreateMarketingRequests)
	require.True(t, profile.CanCreateDaxRequests)
	require.True(t, profile.EmployeeFound)
	require.False(t, profile.RegistrationRequired)
}

func TestProfileServiceGetProfileTreats401AsRegistrationRequired(t *testing.T) {
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{
		err: api.HTTPStatusError{StatusCode: 401},
	})

	profile, err := service.GetProfile(context.Background(), "100245")

	require.NoError(t, err)
	require.Equal(t, "100245", profile.UserID)
	require.False(t, profile.EmployeeFound)
	require.True(t, profile.RegistrationRequired)
	require.Equal(t, "Пользователь не найден в ITILIUM. Требуется регистрация.", profile.StatusMessage)
}

func TestProfileServiceGetProfileTreats403AsPendingRegistration(t *testing.T) {
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{
		err: api.HTTPStatusError{StatusCode: 403},
	})

	profile, err := service.GetProfile(context.Background(), "100245")

	require.NoError(t, err)
	require.Equal(t, "100245", profile.UserID)
	require.False(t, profile.EmployeeFound)
	require.False(t, profile.RegistrationRequired)
	require.True(t, profile.RegistrationPending)
	require.Equal(t, "Ваша заявка на регистрацию еще на рассмотрении.", profile.StatusMessage)
}

func TestProfileServiceGetProfileReturnsUnknownLookupErrors(t *testing.T) {
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{
		err: errors.New("boom"),
	})

	_, err := service.GetProfile(context.Background(), "100245")

	require.Error(t, err)
	require.EqualError(t, err, "boom")
}

func TestProfileServiceRegisterValidatesInput(t *testing.T) {
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{})

	_, err := service.Register(context.Background(), models.RegistrationRequest{})

	require.Error(t, err)
}

func TestProfileServiceRegisterReturnsPendingProfile(t *testing.T) {
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{})

	profile, err := service.Register(context.Background(), models.RegistrationRequest{
		UserID:       "100245",
		FullName:     "Александр Максимов",
		Organization: "ООО Ромашка",
		Department:   "Магазин 17, Казань",
		Position:     "Администратор",
	})

	require.NoError(t, err)
	require.Equal(t, "100245", profile.UserID)
	require.False(t, profile.EmployeeFound)
	require.False(t, profile.RegistrationRequired)
	require.True(t, profile.RegistrationPending)
	require.Equal(t, "Ваша заявка на регистрацию еще на рассмотрении.", profile.StatusMessage)
}

func TestProfileServiceRegisterReturnsExternalError(t *testing.T) {
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{
		regErr: errors.New("upstream registration failed"),
	})

	_, err := service.Register(context.Background(), models.RegistrationRequest{
		UserID:       "100245",
		FullName:     "Александр Максимов",
		Organization: "ООО Ромашка",
		Department:   "Магазин 17, Казань",
		Position:     "Администратор",
	})

	require.EqualError(t, err, "upstream registration failed")
}

func TestProfileServiceGetProfileReturnsStoredPendingProfile(t *testing.T) {
	service := services.NewProfileService(&profileRepositoryStub{
		getProfile: models.UserProfile{
			UserID:               "100245",
			Username:             "100245",
			FullName:             "Заявка на регистрацию отправлена",
			Department:           "Не определено",
			EmployeeFound:        false,
			RegistrationRequired: false,
			RegistrationPending:  true,
			StatusMessage:        "Ваша заявка на регистрацию еще на рассмотрении.",
		},
		getOK: true,
	}, &employeeLookupClientStub{
		err: api.HTTPStatusError{StatusCode: 401},
	})

	profile, err := service.GetProfile(context.Background(), "100245")

	require.NoError(t, err)
	require.True(t, profile.RegistrationPending)
	require.Equal(t, "Ваша заявка на регистрацию еще на рассмотрении.", profile.StatusMessage)
}

func TestProfileServiceFindEmployeeByIdentifierDefaultsAttributeCode(t *testing.T) {
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{
		response: models.EmployeeLookupResult{
			UUID:         "emp-42",
			ServiceCalls: []string{"SC-1", "SC-2"},
		},
	})

	result, err := service.FindEmployeeByIdentifier(context.Background(), models.EmployeeLookupRequest{
		Identifier: "100245",
	})

	require.NoError(t, err)
	require.Equal(t, "100245", result.Identifier)
	require.Equal(t, "id", result.AttributeCode)
	require.Equal(t, "emp-42", result.UUID)
	require.Len(t, result.ServiceCalls, 2)
}
