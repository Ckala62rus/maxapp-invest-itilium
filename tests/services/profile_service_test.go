// Package services_test проверяет ProfileService на уровне публичного поведения.
// Здесь нет реального ITILIUM и базы: внешние зависимости заменены маленькими stubs.
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

// profileRepositoryStub имитирует локальное хранилище профилей.
// Поля структуры задаются прямо в тесте, чтобы получить нужный сценарий.
type profileRepositoryStub struct {
	getProfile models.UserProfile
	getOK      bool
	saved      models.RegistrationRequest
}

// employeeLookupClientStub имитирует клиент ITILIUM для find_employee/registration.
type employeeLookupClientStub struct {
	// response вернётся при успешном FindEmployeeByIdentifier.
	response models.EmployeeLookupResult
	// err позволяет проверить, как сервис реагирует на ошибки find_employee.
	err error
	// regErr позволяет проверить ошибки регистрации.
	regErr error
}

func (s *profileRepositoryStub) GetByUserID(_ context.Context, _ string) (models.UserProfile, bool) {
	// Второе возвращаемое значение bool в Go часто означает "нашлось или нет".
	return s.getProfile, s.getOK
}

func (s *profileRepositoryStub) SaveRegistration(_ context.Context, request models.RegistrationRequest) models.UserProfile {
	// Сохраняем request в stub: при необходимости тест мог бы проверить, что именно было передано.
	s.saved = request
	// Возвращаем профиль в состоянии "ожидает регистрации", как делает реальный repository.
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
	// Если тест настроил ошибку, возвращаем её и не строим успешный ответ.
	if s.err != nil {
		return models.EmployeeLookupResult{}, s.err
	}
	// Если тест не указал AttributeCode, stub повторяет поведение сервиса по умолчанию.
	if s.response.AttributeCode == "" {
		s.response.AttributeCode = request.AttributeCode
	}
	// То же для Identifier: удобно не дублировать его в каждом кейсе.
	if s.response.Identifier == "" {
		s.response.Identifier = request.Identifier
	}

	return s.response, nil
}

func (s *employeeLookupClientStub) RegisterUser(_ context.Context, _ models.RegistrationRequest) error {
	// Для регистрации нам достаточно вернуть nil или заранее заданную ошибку.
	return s.regErr
}

func TestProfileServiceGetProfileFallback(t *testing.T) {
	// Пустой repository + пустой ITILIUM lookup приводят к fallback-профилю.
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{})

	profile, err := service.GetProfile(context.Background(), "100500")

	require.NoError(t, err)
	require.Equal(t, "100500", profile.UserID)
	require.True(t, profile.RegistrationRequired)
}

func TestProfileServiceGetProfileUsesLookupWhenStoredProfileNeedsRegistration(t *testing.T) {
	// Arrange: локально есть старый профиль "нужна регистрация",
	// но ITILIUM уже возвращает полноценного сотрудника.
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

	// Act: запрашиваем профиль.
	profile, err := service.GetProfile(context.Background(), "100245")

	// Assert: сервис должен предпочесть живой lookup из ITILIUM старому fallback-профилю.
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
	// HTTP 401 от find_employee в нашей бизнес-логике значит:
	// "сотрудник не найден, нужно показать регистрацию".
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
	// HTTP 403 от 1С значит: регистрация уже отправлена и ждёт обработки.
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
	// Неизвестные ошибки нельзя маскировать под регистрацию,
	// иначе мы спрячем настоящую техническую проблему.
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{
		err: errors.New("boom"),
	})

	_, err := service.GetProfile(context.Background(), "100245")

	require.Error(t, err)
	require.EqualError(t, err, "boom")
}

func TestProfileServiceRegisterValidatesInput(t *testing.T) {
	// Пустая регистрационная форма должна отвалиться на валидации до вызова ITILIUM.
	service := services.NewProfileService(&profileRepositoryStub{}, &employeeLookupClientStub{})

	_, err := service.Register(context.Background(), models.RegistrationRequest{})

	require.Error(t, err)
}

func TestProfileServiceRegisterReturnsPendingProfile(t *testing.T) {
	// Успешная отправка регистрации переводит пользователя в состояние "pending".
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
	// Если 1С не приняла регистрацию, сервис возвращает эту ошибку вызывающему коду.
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
	// Если локально уже сохранён pending-профиль, а find_employee всё ещё отвечает 401,
	// пользователь должен оставаться на экране ожидания, а не возвращаться к форме регистрации.
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
	// Если UI не передал attributeCode, сервис должен использовать legacy-дефолт "id".
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
