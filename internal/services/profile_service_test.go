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

func TestProfileServiceGetProfileFallback(t *testing.T) {
	service := NewProfileService(&profileRepositoryStub{})

	profile, err := service.GetProfile(context.Background(), "100500")

	require.NoError(t, err)
	require.Equal(t, "100500", profile.UserID)
	require.True(t, profile.RegistrationRequired)
}

func TestProfileServiceRegisterValidatesInput(t *testing.T) {
	service := NewProfileService(&profileRepositoryStub{})

	_, err := service.Register(context.Background(), models.RegistrationRequest{})

	require.Error(t, err)
}
