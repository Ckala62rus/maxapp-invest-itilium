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
	// Пустой идентификатор с MAX недопустим для построения профиля.
	if strings.TrimSpace(userID) == "" {
		return models.UserProfile{}, errors.New("user id is required")
	}

	// Локальный кэш используем только для уже подтверждённого профиля.
	// Статус "заявка на рассмотрении" перепроверяем в ITILIUM: сотрудника могли добавить после отправки анкеты.
	storedProfile, hasStoredProfile := s.repository.GetByUserID(ctx, userID)
	if hasStoredProfile && storedProfile.EmployeeFound && !storedProfile.RegistrationRequired {
		return storedProfile, nil
	}

	// Живой поиск в ITILIUM по MAX user id (если интеграция подключена).
	if s.lookupClient != nil {
		lookup, err := s.FindEmployeeByIdentifier(ctx, models.EmployeeLookupRequest{
			Identifier: userID,
		})
		if err != nil {
			// Ошибка с HTTP-статусом: часть кодов переводим в понятные для UI состояния (регистрация / ожидание).
			if profile, ok := profileFromLookupError(userID, err); ok {
				if hasStoredProfile && storedProfile.RegistrationPending && profile.RegistrationRequired {
					return storedProfile, nil
				}
				return profile, nil
			}
			// Неизвестная ошибка — пробрасываем наверх.
			return models.UserProfile{}, err
		}

		// Ответ find_employee успешно разобран — собираем профиль из полей ITILIUM.
		if profile, ok := profileFromLookup(userID, lookup); ok {
			return profile, nil
		}
	}

	// Нет данных в репозитории, ITILIUM не вернул сотрудника или клиент выключен — шаблон «нужна регистрация».
	return fallbackProfile(userID), nil
}

// Register stores a registration request and returns the newly linked profile snapshot.
func (s *ProfileService) Register(ctx context.Context, request models.RegistrationRequest) (models.UserProfile, error) {
	// Обязательные поля формы: без них ITILIUM не примет заявку.
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

	// Сначала отправляем регистрацию во внешнюю систему; при ошибке локально ничего не фиксируем.
	if err := s.lookupClient.RegisterUser(ctx, request); err != nil {
		return models.UserProfile{}, err
	}

	// Успех: сохраняем в памяти состояние «заявка на рассмотрении» для последующих GET /users/me.
	return s.repository.SaveRegistration(ctx, request), nil
}

// FindEmployeeByIdentifier loads a raw employee payload from ITILIUM for exploration and future identity mapping.
func (s *ProfileService) FindEmployeeByIdentifier(ctx context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error) {
	if strings.TrimSpace(request.Identifier) == "" {
		return models.EmployeeLookupResult{}, errors.New("identifier is required")
	}
	// По умолчанию ищем по полю id (контракт legacy find_employee).
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
	// Имена полей в Raw могут отличаться между версиями 1C — перебираем несколько ключей.
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

	// Если нет ни UUID, ни ФИО/подразделения, ни заявок — считаем, что сотрудника «не нашли».
	found := strings.TrimSpace(lookup.UUID) != "" || fullName != "" || department != "" || len(lookup.ServiceCalls) > 0
	if !found {
		return models.UserProfile{}, false
	}

	// Заполняем обязательные для UI строки дефолтами.
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
		CanCreateMarketingRequests: marketingFlagFromLookup(lookup),
		CanCreateDaxRequests:      daxFlagFromLookup(lookup),
		EmployeeFound:             true,
		RegistrationRequired:      false,
	}, true
}

// marketingFlagFromLookup prefers normalized ITILIUM fields, then re-parses Raw with the same rules as the HTTP client.
func marketingFlagFromLookup(lookup models.EmployeeLookupResult) bool {
	if lookup.CanCreateMarketingRequests {
		return true
	}

	return api.MarketingPermissionFromFindEmployeePayload(lookup.Raw)
}

// daxFlagFromLookup prefers normalized ITILIUM fields, then re-parses Raw for alternate DAX keys.
func daxFlagFromLookup(lookup models.EmployeeLookupResult) bool {
	if lookup.CanCreateDaxRequests {
		return true
	}

	return api.DaxPermissionFromFindEmployeePayload(lookup.Raw)
}

// fallbackProfile keeps the mini app usable when ITILIUM does not know the current user yet.
func fallbackProfile(userID string) models.UserProfile {
	// Минимальный профиль: фронт показывает экран регистрации по флагам ниже.
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
	// Не HTTP-ошибка ITILIUM — отдаём управление выше (сетевая ошибка и т.д.).
	if !errors.As(err, &statusErr) {
		return models.UserProfile{}, false
	}

	switch statusErr.StatusCode {
	case 401, 404:
		// Типичный ответ 1C: пользователь не найден → нужна регистрация.
		return fallbackProfile(userID), true
	case 403:
		// Обычно «заявка уже отправлена и ждёт модерации».
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
		// Прочие коды (5xx и т.д.) не маппим в профиль — пусть уйдут как ошибка.
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
	// Сейчас поддерживаем только строки; остальное игнорируем, чтобы не падать на неожиданных типах.
	switch converted := value.(type) {
	case string:
		return converted
	case nil:
		return ""
	default:
		return ""
	}
}
