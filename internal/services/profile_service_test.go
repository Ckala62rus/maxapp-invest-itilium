package services

import (
	"context"
	"testing"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
	"github.com/stretchr/testify/require"
)

type profileRepositoryStub struct {
	getProfile models.UserProfile
	getOK      bool
	saved      models.RegistrationRequest
}

type employeeLookupClientStub struct {
	response models.EmployeeLookupResult
}

func (s *profileRepositoryStub) GetByUserID(_ context.Context, _ string) (models.UserProfile, bool) {
	return s.getProfile, s.getOK
}

func (s *profileRepositoryStub) SaveRegistration(_ context.Context, request models.RegistrationRequest) models.UserProfile {
	s.saved = request
	return models.UserProfile{
		UserID:               request.UserID,
		FullName:             request.FullName,
		Department:           request.Department,
		EmployeeFound:        true,
		RegistrationRequired: false,
	}
}

func (s *employeeLookupClientStub) FindEmployeeByIdentifier(_ context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error) {
	if s.response.AttributeCode == "" {
		s.response.AttributeCode = request.AttributeCode
	}
	if s.response.Identifier == "" {
		s.response.Identifier = request.Identifier
	}

	return s.response, nil
}

func TestProfileServiceGetProfileFallback(t *testing.T) {
	service := NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{})

	profile, err := service.GetProfile(context.Background(), "100500")

	require.NoError(t, err)
	require.Equal(t, "100500", profile.UserID)
	require.True(t, profile.RegistrationRequired)
}

func TestProfileServiceRegisterValidatesInput(t *testing.T) {
	service := NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{})

	_, err := service.Register(context.Background(), models.RegistrationRequest{})

	require.Error(t, err)
}

func TestProfileServiceFindEmployeeByIdentifierDefaultsAttributeCode(t *testing.T) {
	service := NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{
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
	require.Equal(t, "employee", result.AttributeCode)
	require.Equal(t, "emp-42", result.UUID)
	require.Len(t, result.ServiceCalls, 2)
}
