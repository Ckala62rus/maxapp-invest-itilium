// Package api contains concrete external API clients.
package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/config"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
)

// Client implements the outbound ITILIUM HTTP client.
type Client struct {
	baseURL    string
	login      string
	password   string
	httpClient *http.Client
	logger     *slog.Logger
}

// HTTPStatusError stores an upstream HTTP status so services can branch on 1C workflow states.
type HTTPStatusError struct {
	StatusCode int
}

// Error formats the upstream status into a regular Go error string.
func (e HTTPStatusError) Error() string {
	return "itilium request failed with status " + strconv.Itoa(e.StatusCode)
}

// NewClient creates a real ITILIUM HTTP client.
func NewClient(cfg config.ItiliumConfig, logger *slog.Logger) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	return &Client{
		baseURL:  strings.TrimRight(cfg.BaseURL, "/"),
		login:    cfg.Login,
		password: cfg.Password,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		logger: logger,
	}
}

// ListMyTickets returns the current user's tickets.
func (c *Client) ListMyTickets(ctx context.Context, userID string) ([]models.TicketSummary, error) {
	var response []models.TicketSummary
	if err := c.doJSON(ctx, http.MethodGet, "/tickets", map[string]string{"userId": userID}, nil, &response); err != nil {
		return nil, err
	}

	return response, nil
}

// FindEmployeeByIdentifier loads a user payload from the legacy ITILIUM find_employee endpoint.
func (c *Client) FindEmployeeByIdentifier(ctx context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error) {
	attributeCode := strings.TrimSpace(request.AttributeCode)
	if attributeCode == "" {
		attributeCode = "id"
	}

	form := url.Values{}
	form.Set(attributeCode, request.Identifier)

	var payload map[string]any
	if err := c.doForm(ctx, http.MethodPost, "/find_employee", form, &payload); err != nil {
		return models.EmployeeLookupResult{}, err
	}

	return models.EmployeeLookupResult{
		Identifier:                 request.Identifier,
		AttributeCode:              attributeCode,
		UUID:                       stringFromAny(payload["UUID"]),
		ServiceCalls:               stringSliceFromAny(payload["servicecalls"]),
		CanCreateMarketingRequests: boolFromAny(payload["canCreateMarketingRequests"]),
		CanCreateDaxRequests:       boolFromAny(payload["canCreateDaxRequests"]),
		Raw:                        payload,
	}, nil
}

// RegisterUser sends a registration request to the legacy ITILIUM registration endpoint.
func (c *Client) RegisterUser(ctx context.Context, request models.RegistrationRequest) error {
	form := url.Values{}
	form.Set("id", strings.TrimSpace(request.UserID))
	form.Set("FIO", strings.TrimSpace(request.FullName))
	form.Set("Organization", strings.TrimSpace(request.Organization))
	form.Set("Subdivision", strings.TrimSpace(request.Department))
	form.Set("NamePosition", strings.TrimSpace(request.Position))

	if request.Phone != "" {
		form.Set("Phone", strings.TrimSpace(request.Phone))
	}
	if request.Comment != "" {
		form.Set("Comment", strings.TrimSpace(request.Comment))
	}

	return c.doForm(ctx, http.MethodPost, "/registration", form, nil)
}

// ListResponsibleTickets returns tickets assigned to the current user.
func (c *Client) ListResponsibleTickets(ctx context.Context, userID string) ([]models.TicketSummary, error) {
	var response []models.TicketSummary
	if err := c.doJSON(ctx, http.MethodGet, "/tickets/responsible", map[string]string{"userId": userID}, nil, &response); err != nil {
		return nil, err
	}

	return response, nil
}

// GetTicket returns one detailed ticket card.
func (c *Client) GetTicket(ctx context.Context, userID string, number string) (models.TicketDetail, error) {
	var response models.TicketDetail
	if err := c.doJSON(ctx, http.MethodGet, "/tickets/"+url.PathEscape(number), map[string]string{"userId": userID}, nil, &response); err != nil {
		return models.TicketDetail{}, err
	}

	return response, nil
}

// CreateTicket sends a create request to ITILIUM.
func (c *Client) CreateTicket(ctx context.Context, request models.CreateTicketRequest) (models.TicketDetail, error) {
	var response models.TicketDetail
	if err := c.doJSON(ctx, http.MethodPost, "/tickets", nil, request, &response); err != nil {
		return models.TicketDetail{}, err
	}

	return response, nil
}

// AddComment adds a new comment to a ticket.
func (c *Client) AddComment(ctx context.Context, number string, request models.AddCommentRequest) (models.TicketDetail, error) {
	var response models.TicketDetail
	if err := c.doJSON(ctx, http.MethodPost, "/tickets/"+url.PathEscape(number)+"/comments", nil, request, &response); err != nil {
		return models.TicketDetail{}, err
	}

	return response, nil
}

// ChangeStatus changes the workflow status of a ticket.
func (c *Client) ChangeStatus(ctx context.Context, number string, request models.ChangeStatusRequest) (models.TicketDetail, error) {
	var response models.TicketDetail
	if err := c.doJSON(ctx, http.MethodPost, "/tickets/"+url.PathEscape(number)+"/status", nil, request, &response); err != nil {
		return models.TicketDetail{}, err
	}

	return response, nil
}

// ChangeResponsible changes the responsible person of a ticket.
func (c *Client) ChangeResponsible(ctx context.Context, number string, request models.ChangeResponsibleRequest) (models.TicketDetail, error) {
	var response models.TicketDetail
	if err := c.doJSON(ctx, http.MethodPost, "/tickets/"+url.PathEscape(number)+"/responsible", nil, request, &response); err != nil {
		return models.TicketDetail{}, err
	}

	return response, nil
}

// SearchTicket searches a ticket by number.
func (c *Client) SearchTicket(ctx context.Context, request models.SearchTicketRequest) (models.TicketDetail, error) {
	var response models.TicketDetail
	if err := c.doJSON(ctx, http.MethodPost, "/tickets/search", nil, request, &response); err != nil {
		return models.TicketDetail{}, err
	}

	return response, nil
}

// ListResponsibleOptions returns available assignees for the ticket.
func (c *Client) ListResponsibleOptions(ctx context.Context, userID string, number string) ([]models.ResponsibleOption, error) {
	var response []models.ResponsibleOption
	if err := c.doJSON(ctx, http.MethodGet, "/tickets/"+url.PathEscape(number)+"/responsibles", map[string]string{"userId": userID}, nil, &response); err != nil {
		return nil, err
	}

	return response, nil
}

// doJSON performs a JSON request and logs request and response metadata for troubleshooting.
func (c *Client) doJSON(ctx context.Context, method string, path string, query map[string]string, requestBody any, responseBody any) error {
	if c.baseURL == "" {
		return errors.New("itilium base url is required")
	}

	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("parse itilium url: %w", err)
	}

	values := endpoint.Query()
	for key, value := range query {
		values.Set(key, value)
	}
	endpoint.RawQuery = values.Encode()

	var body io.Reader
	var raw []byte
	if requestBody != nil {
		raw, err = json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewBuffer(raw)
	}

	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	if c.login != "" {
		request.SetBasicAuth(c.login, c.password)
	}

	start := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error("itilium request failed", "method", method, "url", endpoint.String(), "duration_ms", time.Since(start).Milliseconds(), "error", err)
		return fmt.Errorf("perform request: %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	c.logger.Info(
		"itilium request completed",
		"method", method,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	c.logger.Debug(
		"itilium request details",
		"method", method,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"request_body", string(raw),
		"response_body", string(payload),
	)

	if response.StatusCode >= 400 {
		return HTTPStatusError{StatusCode: response.StatusCode}
	}

	if responseBody == nil || len(payload) == 0 {
		return nil
	}

	if err := json.Unmarshal(payload, responseBody); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// doForm performs a x-www-form-urlencoded request for legacy ITILIUM endpoints.
func (c *Client) doForm(ctx context.Context, method string, path string, form url.Values, responseBody any) error {
	if c.baseURL == "" {
		return errors.New("itilium base url is required")
	}

	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("parse itilium url: %w", err)
	}

	encodedForm := form.Encode()
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), strings.NewReader(encodedForm))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.login != "" {
		request.SetBasicAuth(c.login, c.password)
	}

	start := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error("itilium form request failed", "method", method, "url", endpoint.String(), "duration_ms", time.Since(start).Milliseconds(), "error", err)
		return fmt.Errorf("perform form request: %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	c.logger.Info(
		"itilium form request completed",
		"method", method,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	c.logger.Debug(
		"itilium form request details",
		"method", method,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"form_body", encodedForm,
		"response_body", string(payload),
	)

	if response.StatusCode >= 400 {
		return HTTPStatusError{StatusCode: response.StatusCode}
	}

	if responseBody == nil || len(payload) == 0 {
		return nil
	}

	if err := json.Unmarshal(payload, responseBody); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// DemoClient returns deterministic data used by the scaffold and tests.
type DemoClient struct{}

// NewDemoClient creates a deterministic demo client.
func NewDemoClient() *DemoClient {
	return &DemoClient{}
}

// FindEmployeeByIdentifier returns a deterministic employee payload for local learning flows.
func (c *DemoClient) FindEmployeeByIdentifier(_ context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error) {
	attributeCode := strings.TrimSpace(request.AttributeCode)
	if attributeCode == "" {
		attributeCode = "id"
	}

	return models.EmployeeLookupResult{
		Identifier:                 request.Identifier,
		AttributeCode:              attributeCode,
		UUID:                       "demo-employee-uuid",
		ServiceCalls:               []string{"SC-000245", "SC-000244", "SC-000238"},
		CanCreateMarketingRequests: true,
		CanCreateDaxRequests:       true,
		Raw: map[string]any{
			"UUID":                       "demo-employee-uuid",
			"servicecalls":               []string{"SC-000245", "SC-000244", "SC-000238"},
			"canCreateMarketingRequests": true,
			"canCreateDaxRequests":       true,
			"id":                         request.Identifier,
			"displayName":                "Александр Максимов",
		},
	}, nil
}

// RegisterUser accepts registration requests in demo mode without external side effects.
func (c *DemoClient) RegisterUser(_ context.Context, _ models.RegistrationRequest) error {
	return nil
}

// ListMyTickets returns a static list of personal tickets.
func (c *DemoClient) ListMyTickets(_ context.Context, _ string) ([]models.TicketSummary, error) {
	return demoMyTickets(), nil
}

// ListResponsibleTickets returns a static list of responsible tickets.
func (c *DemoClient) ListResponsibleTickets(_ context.Context, _ string) ([]models.TicketSummary, error) {
	return []models.TicketSummary{
		{Number: "SC-000310", Title: "Проблема с авторизацией сотрудника", State: "Ожидает ответа", Deadline: "11.04.2026", ResponsibleTeam: "Отдел ИТ"},
		{Number: "SC-000308", Title: "Нужна смена ответственного", State: "В работе", Deadline: "12.04.2026", ResponsibleTeam: "Отдел ИТ"},
		{Number: "SC-000299", Title: "Добавить комментарий по инциденту", State: "На согласовании", Deadline: "13.04.2026", ResponsibleTeam: "Отдел ИТ"},
	}, nil
}

// GetTicket returns one static ticket card.
func (c *DemoClient) GetTicket(_ context.Context, _ string, number string) (models.TicketDetail, error) {
	return demoTicket(number), nil
}

// CreateTicket creates a synthetic ticket result for the scaffold.
func (c *DemoClient) CreateTicket(_ context.Context, request models.CreateTicketRequest) (models.TicketDetail, error) {
	return models.TicketDetail{
		Number:               "SC-NEW-001",
		Title:                request.Title,
		Description:          request.Description,
		State:                "Зарегистрирована",
		Deadline:             request.ExecutionDate,
		ResponsibleTeam:      "Отдел ИТ",
		CanChangeResponsible: true,
		AvailableStates:      []string{"В работе", "Отложено", "На согласовании"},
		Timeline: []models.CommentEntry{
			{Author: "Система", Message: "Заявка создана и отправлена в ITILIUM.", CreatedAt: time.Now().Format(time.RFC3339)},
		},
	}, nil
}

// AddComment appends a synthetic comment to the static ticket.
func (c *DemoClient) AddComment(_ context.Context, number string, request models.AddCommentRequest) (models.TicketDetail, error) {
	ticket := demoTicket(number)
	ticket.Timeline = append(ticket.Timeline, models.CommentEntry{
		Author:    "Пользователь",
		Message:   request.Message,
		CreatedAt: time.Now().Format(time.RFC3339),
	})

	return ticket, nil
}

// ChangeStatus returns the same demo ticket with a changed state.
func (c *DemoClient) ChangeStatus(_ context.Context, number string, request models.ChangeStatusRequest) (models.TicketDetail, error) {
	ticket := demoTicket(number)
	ticket.State = request.State
	return ticket, nil
}

// ChangeResponsible returns the same demo ticket after synthetic reassignment.
func (c *DemoClient) ChangeResponsible(_ context.Context, number string, _ models.ChangeResponsibleRequest) (models.TicketDetail, error) {
	ticket := demoTicket(number)
	ticket.ResponsibleTeam = "Новый ответственный назначен"
	return ticket, nil
}

// SearchTicket returns the ticket requested by the user.
func (c *DemoClient) SearchTicket(_ context.Context, request models.SearchTicketRequest) (models.TicketDetail, error) {
	return demoTicket(request.Number), nil
}

// ListResponsibleOptions returns deterministic assignee options.
func (c *DemoClient) ListResponsibleOptions(_ context.Context, _ string, _ string) ([]models.ResponsibleOption, error) {
	return []models.ResponsibleOption{
		{Team: "Отдел ИТ", Person: "Иван Петров", Role: "Старший инженер", ExternalID: "emp-1"},
		{Team: "Отдел ИТ", Person: "Елена Орлова", Role: "Системный аналитик", ExternalID: "emp-2"},
		{Team: "Маркетинг", Person: "Мария Соколова", Role: "Маркетолог", ExternalID: "emp-3"},
	}, nil
}

// demoMyTickets centralizes repeated demo ticket list data.
func demoMyTickets() []models.TicketSummary {
	return []models.TicketSummary{
		{Number: "SC-000245", Title: "Не открывается 1С на кассе", State: "В работе", Deadline: "11.04.2026", ResponsibleTeam: "Отдел ИТ"},
		{Number: "SC-000244", Title: "Нужен доступ к отчету по складу", State: "Зарегистрирована", Deadline: "10.04.2026", ResponsibleTeam: "Отдел ИТ"},
		{Number: "SC-000238", Title: "Ошибка печати ценников", State: "На согласовании", Deadline: "09.04.2026", ResponsibleTeam: "Отдел ИТ"},
	}
}

// demoTicket centralizes repeated demo ticket detail data.
func demoTicket(number string) models.TicketDetail {
	return models.TicketDetail{
		Number:               number,
		Title:                "Не открывается 1С на кассе",
		Description:          "После обновления 1С не запускается на рабочем месте кассира.",
		State:                "В работе",
		Deadline:             "11.04.2026",
		ResponsibleTeam:      "Отдел ИТ",
		CanChangeResponsible: true,
		AvailableStates:      []string{"В работе", "Отложено", "На согласовании"},
		Timeline: []models.CommentEntry{
			{Author: "Пользователь", Message: "Не могу открыть 1С на кассе после обновления.", CreatedAt: "2026-04-09T09:15:00+03:00"},
			{Author: "Система", Message: "Заявка зарегистрирована и отправлена в отдел ИТ.", CreatedAt: "2026-04-09T09:18:00+03:00"},
			{Author: "Ответственный", Message: "Проверяю обновление, вернусь с ответом через 10 минут.", CreatedAt: "2026-04-09T09:34:00+03:00"},
		},
	}
}

// stringFromAny converts dynamic JSON values into strings without panicking on unknown types.
func stringFromAny(value any) string {
	if value == nil {
		return ""
	}

	switch converted := value.(type) {
	case string:
		return converted
	case fmt.Stringer:
		return converted.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

// stringSliceFromAny converts raw JSON arrays into a string slice for normalized models.
func stringSliceFromAny(value any) []string {
	switch converted := value.(type) {
	case []string:
		return converted
	case []any:
		result := make([]string, 0, len(converted))
		for _, item := range converted {
			if text := strings.TrimSpace(stringFromAny(item)); text != "" {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

// boolFromAny converts legacy JSON flags into a bool, accepting both boolean and string forms.
func boolFromAny(value any) bool {
	switch converted := value.(type) {
	case bool:
		return converted
	case string:
		return strings.EqualFold(strings.TrimSpace(converted), "true")
	default:
		return false
	}
}
