package services

import (
	"context"
	"testing"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/repository"
	"github.com/stretchr/testify/require"
)

type itiliumClientStub struct {
	getTicketCalls int
}

func (s *itiliumClientStub) FindEmployeeByIdentifier(_ context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error) {
	return models.EmployeeLookupResult{Identifier: request.Identifier, AttributeCode: request.AttributeCode}, nil
}

func (s *itiliumClientStub) ListMyTickets(_ context.Context, _ string) ([]models.TicketSummary, error) {
	return []models.TicketSummary{{Number: "SC-1"}}, nil
}

func (s *itiliumClientStub) ListResponsibleTickets(_ context.Context, _ string) ([]models.TicketSummary, error) {
	return []models.TicketSummary{{Number: "SC-2"}}, nil
}

func (s *itiliumClientStub) GetTicket(_ context.Context, _ string, number string) (models.TicketDetail, error) {
	s.getTicketCalls++
	return models.TicketDetail{Number: number, Title: "demo"}, nil
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

func TestTicketServiceGetTicketUsesCache(t *testing.T) {
	client := &itiliumClientStub{}
	service := NewTicketService(client, repository.NewRedisCache(nil))

	ticket, err := service.GetTicket(context.Background(), "100245", "SC-100")

	require.NoError(t, err)
	require.Equal(t, "SC-100", ticket.Number)
	require.Equal(t, 1, client.getTicketCalls)
}

func TestTicketServiceCreateTicketValidatesInput(t *testing.T) {
	service := NewTicketService(&itiliumClientStub{}, repository.NewRedisCache(nil))

	_, err := service.CreateTicket(context.Background(), models.CreateTicketRequest{})

	require.Error(t, err)
}
