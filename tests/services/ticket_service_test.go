// Package services_test проверяет бизнес-логику сервисов через публичные конструкторы.
// Сервисы получают зависимости через интерфейсы, поэтому в тестах удобно подставлять stubs.
package services_test

import (
	// context.Context почти всегда передаётся в backend-методы Go:
	// через него можно отменять запросы, задавать timeout и прокидывать request_id/user_id.
	"context"
	"testing"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/repository"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/services"
	"github.com/stretchr/testify/require"
)

// itiliumClientStub — тестовая замена реального клиента 1С.
// Он реализует тот же интерфейс, который нужен TicketService,
// но не делает HTTP-запросы и возвращает заранее заданные данные.
type itiliumClientStub struct {
	// Счётчик нужен, чтобы проверить: сервис реально обратился к клиенту ровно один раз.
	getTicketCalls int
}

func (s *itiliumClientStub) FindEmployeeByIdentifier(_ context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error) {
	// Подчёркивание (_) в параметрах значит: аргумент обязателен по сигнатуре,
	// но внутри функции он нам не нужен.
	return models.EmployeeLookupResult{Identifier: request.Identifier, AttributeCode: request.AttributeCode}, nil
}

func (s *itiliumClientStub) ListMyTickets(_ context.Context, _ string) ([]models.TicketSummary, error) {
	// Возвращаем минимальную валидную структуру, достаточную для тестов сервиса.
	return []models.TicketSummary{{Number: "SC-1"}}, nil
}

func (s *itiliumClientStub) ListResponsibleTickets(_ context.Context, _ string) ([]models.TicketSummary, error) {
	return []models.TicketSummary{{Number: "SC-2"}}, nil
}

func (s *itiliumClientStub) GetTicket(_ context.Context, _ string, number string) (models.TicketDetail, error) {
	// Увеличиваем счётчик вызовов, чтобы потом assert-ом проверить поведение.
	s.getTicketCalls++
	return models.TicketDetail{Number: number, Title: "demo"}, nil
}

func (s *itiliumClientStub) ListComments(_ context.Context, _ string, _ string) ([]models.CommentEntry, error) {
	return []models.CommentEntry{{Author: "Тест", Message: "комментарий", CreatedAt: "01.01.2026 12:00:00"}}, nil
}

func (s *itiliumClientStub) CreateTicket(_ context.Context, request models.CreateTicketRequest) (models.TicketDetail, error) {
	return models.TicketDetail{Number: "SC-NEW", Title: request.Title}, nil
}

func (s *itiliumClientStub) AddComment(_ context.Context, number string, _ models.AddCommentRequest) (models.TicketDetail, error) {
	return models.TicketDetail{Number: number}, nil
}

func (s *itiliumClientStub) ChangeStatus(_ context.Context, number string, request models.ChangeStatusRequest) (models.TicketDetail, error) {
	return models.TicketDetail{Number: number, State: request.State}, nil
}

func (s *itiliumClientStub) ChangeResponsible(_ context.Context, number string, _ models.ChangeResponsibleRequest) (models.TicketDetail, error) {
	return models.TicketDetail{Number: number}, nil
}

func (s *itiliumClientStub) SearchTicket(_ context.Context, request models.SearchTicketRequest) (models.TicketDetail, error) {
	return models.TicketDetail{Number: request.Number}, nil
}

func (s *itiliumClientStub) ListResponsibleOptions(_ context.Context, _ string, _ string) ([]models.ResponsibleOption, error) {
	return []models.ResponsibleOption{{ExternalID: "1"}}, nil
}

func (s *itiliumClientStub) ConfirmTicket(_ context.Context, number string, _ models.ConfirmTicketRequest) (models.TicketDetail, error) {
	return models.TicketDetail{Number: number}, nil
}

func (s *itiliumClientStub) ListMarketingServices(_ context.Context, _ string) ([]models.MarketingServiceType, error) {
	return []models.MarketingServiceType{{Code: "design", Name: "Дизайн", FormNumber: "1"}}, nil
}

func (s *itiliumClientStub) ListMarketingSubdivisions(_ context.Context, _ string) ([]models.MarketingSubdivision, error) {
	return []models.MarketingSubdivision{{Name: "Маркетинг"}}, nil
}

func (s *itiliumClientStub) CreateMarketingRequest(_ context.Context, _ models.CreateMarketingRequest) (models.TicketDetail, error) {
	return models.TicketDetail{Number: "SC-M-1", Title: "Marketing"}, nil
}

func TestTicketServiceGetTicketUsesCache(t *testing.T) {
	// Arrange: создаём stub и сервис. Redis cache передаём с nil-клиентом,
	// потому что в этом unit-тесте нам не нужен настоящий Redis.
	client := &itiliumClientStub{}
	service := services.NewTicketService(client, repository.NewRedisCache(nil))

	// Act: вызываем метод, который хотим проверить.
	ticket, err := service.GetTicket(context.Background(), "100245", "SC-100")

	// Assert: require.NoError остановит тест, если err != nil.
	require.NoError(t, err)
	require.Equal(t, "SC-100", ticket.Number)
	// Проверяем, что сервис сходил в клиент за карточкой.
	require.Equal(t, 1, client.getTicketCalls)
}

func TestTicketServiceCreateTicketValidatesInput(t *testing.T) {
	// Пустой CreateTicketRequest должен быть отклонён сервисной валидацией.
	service := services.NewTicketService(&itiliumClientStub{}, repository.NewRedisCache(nil))

	_, err := service.CreateTicket(context.Background(), models.CreateTicketRequest{})

	// Здесь нам важен сам факт ошибки, а не точный текст.
	require.Error(t, err)
}

func TestTicketServiceChangeStatusPostponedRequiresCommentAndDate(t *testing.T) {
	service := services.NewTicketService(&itiliumClientStub{}, repository.NewRedisCache(nil))

	_, err := service.ChangeStatus(context.Background(), "SC-1", models.ChangeStatusRequest{
		UserID: "1",
		State:  "05_Отложено",
	})
	require.Error(t, err)

	_, err = service.ChangeStatus(context.Background(), "SC-1", models.ChangeStatusRequest{
		UserID:  "1",
		State:   "05_Отложено",
		Comment: "перенос",
	})
	require.Error(t, err)

	ticket, err := service.ChangeStatus(context.Background(), "SC-1", models.ChangeStatusRequest{
		UserID:  "1",
		State:   "05_Отложено",
		Comment: "перенос",
		Date:    "2026-07-01",
	})
	require.NoError(t, err)
	require.Equal(t, "05_Отложено", ticket.State)
}

func TestTicketServiceAddCommentRequiresMessageOrAttachment(t *testing.T) {
	// Комментарий без текста и без вложений не имеет смысла для 1С,
	// поэтому сервис должен вернуть ошибку до вызова ITILIUM.
	service := services.NewTicketService(&itiliumClientStub{}, repository.NewRedisCache(nil))

	_, err := service.AddComment(context.Background(), "SC-1", models.AddCommentRequest{})

	require.Error(t, err)
}

func TestTicketServiceListComments(t *testing.T) {
	service := services.NewTicketService(&itiliumClientStub{}, repository.NewRedisCache(nil))

	comments, err := service.ListComments(context.Background(), "40367639", "0000019683")

	require.NoError(t, err)
	require.Len(t, comments, 1)
	require.Equal(t, "комментарий", comments[0].Message)
}
