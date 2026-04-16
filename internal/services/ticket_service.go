package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/repository"
)

// ItiliumClient describes external ITILIUM calls consumed by the service layer.
type ItiliumClient interface {
	// FindEmployeeByIdentifier requests a raw employee payload from ITILIUM.
	FindEmployeeByIdentifier(ctx context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error)
	// ListMyTickets returns tickets created by the current user.
	ListMyTickets(ctx context.Context, userID string) ([]models.TicketSummary, error)
	// ListResponsibleTickets returns tickets assigned to the current user.
	ListResponsibleTickets(ctx context.Context, userID string) ([]models.TicketSummary, error)
	// GetTicket returns a detailed ticket card.
	GetTicket(ctx context.Context, userID string, number string) (models.TicketDetail, error)
	// CreateTicket creates a new ITILIUM ticket.
	CreateTicket(ctx context.Context, request models.CreateTicketRequest) (models.TicketDetail, error)
	// AddComment adds a new timeline entry to a ticket.
	AddComment(ctx context.Context, number string, request models.AddCommentRequest) (models.TicketDetail, error)
	// ChangeStatus updates the ticket workflow state.
	ChangeStatus(ctx context.Context, number string, request models.ChangeStatusRequest) (models.TicketDetail, error)
	// ChangeResponsible changes the current responsible person.
	ChangeResponsible(ctx context.Context, number string, request models.ChangeResponsibleRequest) (models.TicketDetail, error)
	// SearchTicket resolves a ticket number into the full detail model.
	SearchTicket(ctx context.Context, request models.SearchTicketRequest) (models.TicketDetail, error)
	// ListResponsibleOptions returns available responsible persons for the ticket.
	ListResponsibleOptions(ctx context.Context, userID string, number string) ([]models.ResponsibleOption, error)
}

// TicketService orchestrates ticket list and workflow use cases.
type TicketService struct {
	client ItiliumClient
	cache  *repository.RedisCache
}

// NewTicketService creates a new ticket service.
func NewTicketService(client ItiliumClient, cache *repository.RedisCache) *TicketService {
	return &TicketService{
		client: client,
		cache:  cache,
	}
}

// ListMyTickets returns the current user's own tickets.
func (s *TicketService) ListMyTickets(ctx context.Context, userID string) ([]models.TicketSummary, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user id is required")
	}

	// Проксируем в ITILIUM без локальной логики.
	return s.client.ListMyTickets(ctx, userID)
}

// ListResponsibleTickets returns tickets assigned to the current user.
func (s *TicketService) ListResponsibleTickets(ctx context.Context, userID string) ([]models.TicketSummary, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user id is required")
	}

	// Список заявок, где пользователь указан ответственным.
	return s.client.ListResponsibleTickets(ctx, userID)
}

// GetTicket returns one ticket detail and caches it for repeated reads.
func (s *TicketService) GetTicket(ctx context.Context, userID string, number string) (models.TicketDetail, error) {
	if strings.TrimSpace(number) == "" {
		return models.TicketDetail{}, errors.New("ticket number is required")
	}

	cacheKey := fmt.Sprintf("ticket:%s:%s", userID, number)

	// Сначала пробуем Redis (если включён): карточка заявки меняется реже, чем её открывают.
	if s.cache != nil {
		var cached models.TicketDetail
		ok, err := s.cache.GetJSON(ctx, cacheKey, &cached)
		if err == nil && ok {
			return cached, nil
		}
	}

	ticket, err := s.client.GetTicket(ctx, userID, number)
	if err != nil {
		return models.TicketDetail{}, err
	}

	if s.cache != nil {
		// Кэшируем ненадолго: статус и комментарии должны подтянуться с новым запросом.
		_ = s.cache.SetJSON(ctx, cacheKey, ticket, 5*time.Minute)
	}

	return ticket, nil
}

// CreateTicket creates a new ITILIUM ticket.
func (s *TicketService) CreateTicket(ctx context.Context, request models.CreateTicketRequest) (models.TicketDetail, error) {
	if strings.TrimSpace(request.UserID) == "" {
		return models.TicketDetail{}, errors.New("user id is required")
	}
	if strings.TrimSpace(request.Title) == "" {
		return models.TicketDetail{}, errors.New("title is required")
	}
	if strings.TrimSpace(request.Description) == "" {
		return models.TicketDetail{}, errors.New("description is required")
	}

	// Создание всегда идёт в ITILIUM; кэш карточки здесь не инвалидируем (ещё нет стабильного номера до ответа).
	return s.client.CreateTicket(ctx, request)
}

// SearchTicket finds a ticket by its number.
func (s *TicketService) SearchTicket(ctx context.Context, request models.SearchTicketRequest) (models.TicketDetail, error) {
	if strings.TrimSpace(request.Number) == "" {
		return models.TicketDetail{}, errors.New("ticket number is required")
	}

	return s.client.SearchTicket(ctx, request)
}

// AddComment appends a new comment to a ticket.
func (s *TicketService) AddComment(ctx context.Context, number string, request models.AddCommentRequest) (models.TicketDetail, error) {
	if strings.TrimSpace(number) == "" {
		return models.TicketDetail{}, errors.New("ticket number is required")
	}
	if strings.TrimSpace(request.Message) == "" {
		return models.TicketDetail{}, errors.New("message is required")
	}

	// Ответ — полная карточка; кэш по этому номеру устареет по TTL сам.
	return s.client.AddComment(ctx, number, request)
}

// ChangeStatus performs a workflow transition.
func (s *TicketService) ChangeStatus(ctx context.Context, number string, request models.ChangeStatusRequest) (models.TicketDetail, error) {
	if strings.TrimSpace(number) == "" {
		return models.TicketDetail{}, errors.New("ticket number is required")
	}
	if strings.TrimSpace(request.State) == "" {
		return models.TicketDetail{}, errors.New("state is required")
	}

	return s.client.ChangeStatus(ctx, number, request)
}

// ChangeResponsible assigns the ticket to a new person.
func (s *TicketService) ChangeResponsible(ctx context.Context, number string, request models.ChangeResponsibleRequest) (models.TicketDetail, error) {
	if strings.TrimSpace(number) == "" {
		return models.TicketDetail{}, errors.New("ticket number is required")
	}
	if strings.TrimSpace(request.ResponsibleID) == "" {
		return models.TicketDetail{}, errors.New("responsible id is required")
	}

	// Смена ответственного в ITILIUM; возвращается обновлённая карточка.
	return s.client.ChangeResponsible(ctx, number, request)
}

// ListResponsibleOptions returns available assignees for the ticket.
func (s *TicketService) ListResponsibleOptions(ctx context.Context, userID string, number string) ([]models.ResponsibleOption, error) {
	if strings.TrimSpace(number) == "" {
		return nil, errors.New("ticket number is required")
	}

	// Справочник возможных исполнителей для UI (выпадающий список).
	return s.client.ListResponsibleOptions(ctx, userID, number)
}
