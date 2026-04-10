// Package repository contains infrastructure implementations for local state.
package repository

import (
	"context"
	"sync"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
)

// MemoryUserRepository stores user profile snapshots in memory for the demo scaffold.
type MemoryUserRepository struct {
	mu      sync.RWMutex
	profile map[string]models.UserProfile
}

// NewMemoryUserRepository creates a repository with demo profile data.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		profile: map[string]models.UserProfile{
			"100245": {
				UserID:               "100245",
				Username:             "amaximov",
				FullName:             "Александр Максимов",
				Department:           "Магазин 17, Казань",
				EmployeeFound:        false,
				RegistrationRequired: true,
			},
		},
	}
}

// GetByUserID returns a profile snapshot by MAX user id.
func (r *MemoryUserRepository) GetByUserID(_ context.Context, userID string) (models.UserProfile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	profile, ok := r.profile[userID]
	return profile, ok
}

// SaveRegistration stores a synthetic registered profile for demo flows.
func (r *MemoryUserRepository) SaveRegistration(_ context.Context, request models.RegistrationRequest) models.UserProfile {
	r.mu.Lock()
	defer r.mu.Unlock()

	profile := models.UserProfile{
		UserID:               request.UserID,
		Username:             request.UserID,
		FullName:             request.FullName,
		Department:           request.Department,
		EmployeeFound:        true,
		RegistrationRequired: false,
	}

	r.profile[request.UserID] = profile

	return profile
}
