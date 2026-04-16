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

// NewMemoryUserRepository creates an in-memory profile store for registration flows.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		profile: map[string]models.UserProfile{},
	}
}

// GetByUserID returns a profile snapshot by MAX user id.
func (r *MemoryUserRepository) GetByUserID(_ context.Context, userID string) (models.UserProfile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	profile, ok := r.profile[userID]
	return profile, ok
}

// SaveRegistration stores a local pending-review state for registration flows.
func (r *MemoryUserRepository) SaveRegistration(_ context.Context, request models.RegistrationRequest) models.UserProfile {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Персистентного БД-профиля пока нет: держим в памяти флаг «ожидает рассмотрения» для GET /users/me.
	profile := models.UserProfile{
		UserID:               request.UserID,
		Username:             request.UserID,
		FullName:             request.FullName,
		Department:           request.Department,
		EmployeeFound:        false,
		RegistrationRequired: false,
		RegistrationPending:  true,
		StatusMessage:        "Ваша заявка на регистрацию еще на рассмотрении.",
	}

	r.profile[request.UserID] = profile

	return profile
}
