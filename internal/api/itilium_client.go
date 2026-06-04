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
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/config"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/middleware"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
)

const (
	responsibleTicketHydrationConcurrency  = 6
	responsibleTicketHydrationTimeout      = 5 * time.Second
	responsibleTicketHydrationTotalTimeout = 12 * time.Second

	// list_sc на тестовом контуре падает с УстановитьСтроку, если sc_number слишком длинный (много номеров через ';').
	listSCMaxNumbersPerRequest = 15

	myTicketHydrationConcurrency  = 6
	myTicketHydrationTimeout      = 5 * time.Second
	myTicketHydrationTotalTimeout = 10 * time.Second
	myTicketHydrationMaxTickets   = 30
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
	lookup, err := c.FindEmployeeByIdentifier(ctx, models.EmployeeLookupRequest{
		Identifier:    userID,
		AttributeCode: "id",
	})
	if err != nil {
		return nil, err
	}

	numbers := normalizeServiceCallNumbers(lookup.ServiceCalls)
	if len(numbers) == 0 {
		return []models.TicketSummary{}, nil
	}

	summaryFromProfileNumbers := func() []models.TicketSummary {
		out := make([]models.TicketSummary, 0, len(numbers))
		for _, number := range numbers {
			out = append(out, models.TicketSummary{
				Number: number,
				Title:  "Заявка " + number,
				State:  "Откройте карточку",
			})
		}
		return out
	}

	hydrateFromFindSC := func(summaries []models.TicketSummary) []models.TicketSummary {
		return c.hydrateTicketSummariesFromFindSC(ctx, userID, summaries)
	}

	// list_sc: id + sc_number (номера через ';'). Длинная строка ломает 1С — шлём пачками.
	response, err := c.fetchListSCSummaries(ctx, userID, numbers)
	if err != nil {
		c.logger.Warn(
			"list_sc failed, returning my tickets from servicecalls only",
			"error", err,
			"user_id", userID,
			"ticket_count", len(numbers),
			"request_id", middleware.RequestIDFromContext(ctx),
		)
		return hydrateFromFindSC(summaryFromProfileNumbers()), nil
	}

	if len(response) == 0 {
		return hydrateFromFindSC(summaryFromProfileNumbers()), nil
	}

	return hydrateFromFindSC(response), nil
}

// fetchListSCSummaries запрашивает list_sc пачками, чтобы не превышать лимит длины sc_number в 1С.
func (c *Client) fetchListSCSummaries(ctx context.Context, userID string, numbers []string) ([]models.TicketSummary, error) {
	if len(numbers) == 0 {
		return []models.TicketSummary{}, nil
	}

	merged := make([]models.TicketSummary, 0, len(numbers))
	for start := 0; start < len(numbers); start += listSCMaxNumbersPerRequest {
		end := start + listSCMaxNumbersPerRequest
		if end > len(numbers) {
			end = len(numbers)
		}

		form := url.Values{}
		form.Set("id", strings.TrimSpace(userID))
		form.Set("sc_number", strings.Join(numbers[start:end], ";"))

		payload, err := c.doMultipartFormPostBytes(ctx, "/list_sc", form)
		if err != nil {
			return nil, err
		}

		batch, err := parseListSCResponse(payload)
		if err != nil {
			return nil, err
		}
		merged = append(merged, batch...)
	}

	return merged, nil
}

func normalizeServiceCallNumbers(serviceCalls []string) []string {
	uniq := make(map[string]struct{}, len(serviceCalls))
	result := make([]string, 0, len(serviceCalls))
	for _, item := range serviceCalls {
		number := strings.TrimSpace(item)
		if number == "" {
			continue
		}
		if _, exists := uniq[number]; exists {
			continue
		}
		uniq[number] = struct{}{}
		result = append(result, number)
	}
	return result
}

func parseListSCResponse(payload []byte) ([]models.TicketSummary, error) {
	clean := bytes.TrimPrefix(payload, []byte{0xEF, 0xBB, 0xBF})
	rawText := strings.TrimSpace(string(clean))
	if rawText == "" {
		return []models.TicketSummary{}, nil
	}

	// Некоторые legacy-методы (например list_sc_responsible) возвращают просто массив номеров.
	var numberList []string
	if err := json.Unmarshal(clean, &numberList); err == nil {
		result := make([]models.TicketSummary, 0, len(numberList))
		for _, number := range numberList {
			trimmed := strings.TrimSpace(number)
			if trimmed == "" {
				continue
			}
			result = append(result, models.TicketSummary{
				Number: trimmed,
				Title:  "Заявка " + trimmed,
				State:  "Откройте карточку",
			})
		}
		return result, nil
	}

	var list []map[string]any
	if err := json.Unmarshal(clean, &list); err == nil {
		return mapListSCItems(list), nil
	}

	var envelope map[string]any
	if err := json.Unmarshal(clean, &envelope); err != nil {
		return nil, fmt.Errorf("decode list_sc response: %w", err)
	}

	for _, key := range []string{"data", "items", "list", "result"} {
		value, ok := envelope[key]
		if !ok {
			continue
		}

		items, ok := value.([]any)
		if !ok {
			continue
		}

		normalized := make([]map[string]any, 0, len(items))
		for _, item := range items {
			if m, castOK := item.(map[string]any); castOK {
				normalized = append(normalized, m)
			}
		}
		return mapListSCItems(normalized), nil
	}

	return []models.TicketSummary{}, nil
}

func mapListSCItems(items []map[string]any) []models.TicketSummary {
	result := make([]models.TicketSummary, 0, len(items))
	for _, item := range items {
		number := pickStringFromMap(item, "number", "Number", "sc_number", "sc", "ServiceCallNumber", "Номер")
		if number == "" {
			continue
		}

		result = append(result, models.TicketSummary{
			Number:          number,
			Title:           pickStringFromMap(item, "title", "Title", "shortDescription", "ShortDescription", "Тема"),
			State:           pickStringFromMap(item, "state", "State", "status", "Status", "Состояние"),
			CreationDate:    pickStringFromMap(item, "creationDate", "CreationDate", "dateCreate", "createdAt", "ДатаСоздания"),
			Deadline:        pickStringFromMap(item, "deadline", "Deadline", "executionDate", "ДатаИсполнения"),
			ResponsibleTeam: pickStringFromMap(item, "responsibleTeam", "ResponsibleTeam", "client", "OU", "Подразделение"),
		})
	}
	return result
}

func (c *Client) hydrateTicketSummariesFromFindSC(ctx context.Context, userID string, summaries []models.TicketSummary) []models.TicketSummary {
	if len(summaries) == 0 {
		return summaries
	}

	hydrateCount := len(summaries)
	if hydrateCount > myTicketHydrationMaxTickets {
		hydrateCount = myTicketHydrationMaxTickets
	}

	hydrationCtx, cancelHydration := context.WithTimeout(ctx, myTicketHydrationTotalTimeout)
	defer cancelHydration()

	out := make([]models.TicketSummary, len(summaries))
	copy(out, summaries)

	sem := make(chan struct{}, myTicketHydrationConcurrency)
	var wg sync.WaitGroup

	for index := 0; index < hydrateCount; index++ {
		summary := summaries[index]
		number := strings.TrimSpace(summary.Number)
		if number == "" || ticketSummaryHasListFields(summary) {
			continue
		}

		wg.Add(1)
		go func(index int, number string, fallback models.TicketSummary) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-hydrationCtx.Done():
				return
			}

			ticketCtx, cancel := context.WithTimeout(hydrationCtx, myTicketHydrationTimeout)
			defer cancel()

			detail, err := c.GetTicket(ticketCtx, userID, number)
			if err != nil {
				c.logger.Warn(
					"find_sc failed while hydrating my ticket summary",
					"error", err,
					"user_id", userID,
					"ticket_number", number,
					"request_id", middleware.RequestIDFromContext(ctx),
				)
				return
			}

			out[index] = ticketSummaryFromDetail(fallback, detail)
		}(index, number, summary)
	}
	wg.Wait()

	if errors.Is(hydrationCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
		c.logger.Warn(
			"my ticket summary hydration timed out",
			"ticket_count", len(summaries),
			"hydrated_limit", hydrateCount,
			"timeout_ms", myTicketHydrationTotalTimeout.Milliseconds(),
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", userID,
		)
	}

	return out
}

func ticketSummaryHasListFields(summary models.TicketSummary) bool {
	title := strings.TrimSpace(summary.Title)
	state := strings.TrimSpace(summary.State)
	return title != "" &&
		!strings.HasPrefix(title, "Заявка ") &&
		state != "" &&
		state != "Откройте карточку" &&
		strings.TrimSpace(summary.CreationDate) != ""
}

func ticketSummaryFromDetail(fallback models.TicketSummary, detail models.TicketDetail) models.TicketSummary {
	summary := fallback
	if strings.TrimSpace(detail.Number) != "" {
		summary.Number = detail.Number
	}
	if strings.TrimSpace(detail.Title) != "" {
		summary.Title = detail.Title
	}
	if strings.TrimSpace(detail.State) != "" {
		summary.State = detail.State
	}
	if strings.TrimSpace(detail.CreationDate) != "" {
		summary.CreationDate = detail.CreationDate
	}
	if strings.TrimSpace(detail.Deadline) != "" {
		summary.Deadline = detail.Deadline
	}
	if strings.TrimSpace(detail.ResponsibleTeam) != "" {
		summary.ResponsibleTeam = detail.ResponsibleTeam
	}
	return summary
}

// FindEmployeeByIdentifier loads a user payload from the legacy ITILIUM find_employee endpoint.
func (c *Client) FindEmployeeByIdentifier(ctx context.Context, request models.EmployeeLookupRequest) (models.EmployeeLookupResult, error) {
	attributeCode := strings.TrimSpace(request.AttributeCode)
	if attributeCode == "" {
		attributeCode = "id"
	}

	// Legacy: одно поле формы — искомое значение (например id=<MAX user id>).
	form := url.Values{}
	form.Set(attributeCode, request.Identifier)

	var payload map[string]any
	if err := c.doForm(ctx, http.MethodPost, "/find_employee", form, &payload); err != nil {
		return models.EmployeeLookupResult{}, err
	}

	// Флаги прав: имена ключей различаются по контурам 1С (латиница/PascalCase/русские имена обёрток).
	marketing := MarketingPermissionFromFindEmployeePayload(payload)
	if !marketing {
		marketing = boolFromAny(firstAny(payload, "canCreateMarketingRequests", "canCreateMarketing", "can_marketing"))
	}
	dax := DaxPermissionFromFindEmployeePayload(payload)
	if !dax {
		dax = boolFromAny(firstAny(payload, "canCreateDaxRequests", "canCreateDax", "can_dax"))
	}

	return models.EmployeeLookupResult{
		Identifier:                 request.Identifier,
		AttributeCode:              attributeCode,
		UUID:                       stringFromAny(payload["UUID"]),
		ServiceCalls:               stringSliceFromAny(payload["servicecalls"]),
		CanCreateMarketingRequests: marketing,
		CanCreateDaxRequests:       dax,
		Raw:                        payload,
	}, nil
}

// RegisterUser sends a registration request to the legacy ITILIUM registration endpoint.
func (c *Client) RegisterUser(ctx context.Context, request models.RegistrationRequest) error {
	// Имена полей заданы контрактом 1C (FIO, Subdivision, NamePosition, …).
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
	// Legacy-контракт 1С: list_sc_responsible с полем id (MAX user id).
	form := url.Values{}
	form.Set("id", strings.TrimSpace(userID))

	// По данным Postman этот метод чувствителен к multipart/form-data.
	payload, err := c.doMultipartFormPostBytes(ctx, "/list_sc_responsible", form)
	if err != nil {
		return nil, err
	}

	numbers, err := parseTicketNumbers(payload)
	if err != nil {
		return nil, err
	}
	if len(numbers) == 0 {
		return []models.TicketSummary{}, nil
	}

	// По уточненному контракту карточки из ответственности запрашиваем через find_sc для каждого номера.
	// Ограничиваем и одну карточку, и всю пачку: при большом списке backend должен успеть отдать частичный результат до HTTP write_timeout.
	hydrationCtx, cancelHydration := context.WithTimeout(ctx, responsibleTicketHydrationTotalTimeout)
	defer cancelHydration()

	summaries := make([]models.TicketSummary, len(numbers))
	sem := make(chan struct{}, responsibleTicketHydrationConcurrency)
	var wg sync.WaitGroup

	for index, number := range numbers {
		summaries[index] = models.TicketSummary{
			Number: number,
			Title:  "Заявка " + number,
			State:  "Откройте карточку",
		}

		wg.Add(1)
		go func(index int, number string) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-hydrationCtx.Done():
				return
			}

			ticketCtx, cancel := context.WithTimeout(hydrationCtx, responsibleTicketHydrationTimeout)
			defer cancel()

			detail, err := c.GetTicket(ticketCtx, userID, number)
			if err != nil {
				// Отдельная карточка может быть недоступна — в списке оставляем номер без падения всего запроса.
				c.logger.Warn(
					"find_sc failed while hydrating responsible ticket summary",
					"number", number,
					"error", err,
					"request_id", middleware.RequestIDFromContext(ctx),
					"user_id", middleware.UserIDFromContext(ctx),
				)
				return
			}

			summaries[index] = models.TicketSummary{
				Number:          detail.Number,
				Title:           detail.Title,
				State:           detail.State,
				CreationDate:    detail.CreationDate,
				Deadline:        detail.Deadline,
				ResponsibleTeam: detail.ResponsibleTeam,
			}
		}(index, number)
	}
	wg.Wait()

	if errors.Is(hydrationCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
		c.logger.Warn(
			"responsible ticket summary hydration timed out",
			"ticket_count", len(numbers),
			"timeout_ms", responsibleTicketHydrationTotalTimeout.Milliseconds(),
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
	}

	return summaries, nil
}

func (c *Client) doMultipartFormPostBytes(ctx context.Context, path string, form url.Values) ([]byte, error) {
	return c.doMultipartFormPostBytesWithFiles(ctx, path, form, nil)
}

func (c *Client) doMultipartFormPostBytesWithFiles(ctx context.Context, path string, form url.Values, files []models.FileAttachment) ([]byte, error) {
	if c.baseURL == "" {
		return nil, errors.New("itilium base url is required")
	}

	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse itilium url: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, values := range form {
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				return nil, fmt.Errorf("write multipart field %q: %w", key, err)
			}
		}
	}
	for _, fa := range files {
		part, err := writer.CreateFormFile("files", fa.Filename)
		if err != nil {
			return nil, fmt.Errorf("create form file: %w", err)
		}
		if _, err := part.Write(fa.Data); err != nil {
			return nil, fmt.Errorf("write form file: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if c.login != "" {
		request.SetBasicAuth(c.login, c.password)
	}

	start := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error(
			"itilium multipart form request failed",
			"method", http.MethodPost,
			"url", endpoint.String(),
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
		return nil, fmt.Errorf("perform multipart form request: %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return c.logMultipartFormResponse(ctx, endpoint, form, files, start, response.StatusCode, payload)
}

func (c *Client) doPostQueryWithMultipartForm(ctx context.Context, path string, query, form url.Values, files []models.FileAttachment) ([]byte, error) {
	if c.baseURL == "" {
		return nil, errors.New("itilium base url is required")
	}

	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse itilium url: %w", err)
	}
	endpoint.RawQuery = query.Encode()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, values := range form {
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				return nil, fmt.Errorf("write multipart field %q: %w", key, err)
			}
		}
	}
	for _, fa := range files {
		part, err := writer.CreateFormFile("files", fa.Filename)
		if err != nil {
			return nil, fmt.Errorf("create form file: %w", err)
		}
		if _, err := part.Write(fa.Data); err != nil {
			return nil, fmt.Errorf("write form file: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if c.login != "" {
		request.SetBasicAuth(c.login, c.password)
	}

	start := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error(
			"itilium query multipart post request failed",
			"method", http.MethodPost,
			"url", endpoint.String(),
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
		return nil, fmt.Errorf("perform query multipart post request: %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	respText := strings.TrimPrefix(string(payload), "\ufeff")
	c.logger.Info(
		"itilium query multipart post request completed",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"query_string", truncateLogString(query.Encode(), 12000),
		"form_fields", truncateLogString(form.Encode(), 12000),
		"response_body", truncateLogString(respText, 8000),
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)
	c.logger.Debug(
		"itilium query multipart post request details",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"query", query.Encode(),
		"form_fields", form.Encode(),
		"response_body", respText,
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)

	if response.StatusCode >= 400 {
		return nil, HTTPStatusError{StatusCode: response.StatusCode}
	}

	return payload, nil
}

func (c *Client) logMultipartFormResponse(ctx context.Context, endpoint *url.URL, form url.Values, files []models.FileAttachment, start time.Time, statusCode int, payload []byte) ([]byte, error) {
	respText := strings.TrimPrefix(string(payload), "\ufeff")
	c.logger.Info(
		"itilium multipart form request completed",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", statusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"form_fields", truncateLogString(form.Encode(), 12000),
		"files", multipartFileLogMeta(files),
		"response_body", truncateLogString(respText, 8000),
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)
	c.logger.Debug(
		"itilium multipart form request details",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", statusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"form_fields", form.Encode(),
		"files", multipartFileLogMeta(files),
		"response_body", respText,
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)

	if statusCode >= 400 {
		return nil, HTTPStatusError{StatusCode: statusCode}
	}

	return payload, nil
}

func multipartFileLogMeta(files []models.FileAttachment) []string {
	out := make([]string, 0, len(files))
	for _, fa := range files {
		out = append(out, fmt.Sprintf("%s:%dB", fa.Filename, len(fa.Data)))
	}
	return out
}

// doPostEmptyQuery выполняет POST без тела: legacy confirm_sc и fallback responsibles_sc с параметрами в query.
func (c *Client) doPostEmptyQuery(ctx context.Context, path string, query url.Values) ([]byte, error) {
	if c.baseURL == "" {
		return nil, errors.New("itilium base url is required")
	}

	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse itilium url: %w", err)
	}
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if c.login != "" {
		request.SetBasicAuth(c.login, c.password)
	}

	start := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error(
			"itilium empty post request failed",
			"method", http.MethodPost,
			"url", endpoint.String(),
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
		return nil, fmt.Errorf("perform post request: %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	respText := strings.TrimPrefix(string(payload), "\ufeff")
	c.logger.Info(
		"itilium empty post request completed",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"response_body", truncateLogString(respText, 8000),
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)

	if response.StatusCode >= 400 {
		return nil, HTTPStatusError{StatusCode: response.StatusCode}
	}

	return payload, nil
}

// doPostQueryWithUrlencodedBody отправляет POST с полями в query и датой/доп. полями в теле form-urlencoded.
func (c *Client) doPostQueryWithUrlencodedBody(ctx context.Context, path string, query, body url.Values) ([]byte, error) {
	if c.baseURL == "" {
		return nil, errors.New("itilium base url is required")
	}

	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse itilium url: %w", err)
	}
	endpoint.RawQuery = query.Encode()

	encodedBody := body.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(encodedBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.login != "" {
		request.SetBasicAuth(c.login, c.password)
	}

	start := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error(
			"itilium hybrid post request failed",
			"method", http.MethodPost,
			"url", endpoint.String(),
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
		return nil, fmt.Errorf("perform post request: %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	respText := strings.TrimPrefix(string(payload), "\ufeff")
	c.logger.Info(
		"itilium hybrid post request completed",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"query_string", truncateLogString(query.Encode(), 12000),
		"urlencoded_body", truncateLogString(encodedBody, 12000),
		"response_body", truncateLogString(respText, 8000),
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)
	c.logger.Debug(
		"itilium hybrid post request details",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"query", query.Encode(),
		"form_body", encodedBody,
		"response_body", respText,
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)

	if response.StatusCode >= 400 {
		return nil, HTTPStatusError{StatusCode: response.StatusCode}
	}

	return payload, nil
}

// parseResponsiblesScResponse разбирает ответ responsibles_sc: либо массив команд с вложенным responsibles, либо плоский список.
func parseResponsiblesScResponse(payloadRaw []byte) ([]models.ResponsibleOption, error) {
	clean := bytes.TrimPrefix(payloadRaw, []byte{0xEF, 0xBB, 0xBF})
	var rows []map[string]any
	if err := json.Unmarshal(clean, &rows); err != nil {
		return nil, fmt.Errorf("decode responsibles_sc response: %w", err)
	}

	result := make([]models.ResponsibleOption, 0, len(rows)*4)
	for _, item := range rows {
		team := pickStringFromMap(item, "responsibleTeamTitle", "ResponsibleTeamTitle", "team", "Team")
		teamID := pickStringFromMap(item, "responsibleTeamId", "ResponsibleTeamId", "teamId", "TeamId")

		if nested, ok := item["responsibles"]; ok {
			arr, ok := nested.([]any)
			if !ok {
				continue
			}
			for _, emp := range arr {
				m, ok := emp.(map[string]any)
				if !ok {
					continue
				}
				externalID := pickStringFromMap(m, "responsibleEmployeeId", "ResponsibleEmployeeId", "externalId", "ExternalID", "employeeId", "id")
				person := pickStringFromMap(m, "responsibleEmployeeTitle", "ResponsibleEmployeeTitle", "person", "Person", "title", "name", "fio")
				if externalID == "" && person == "" {
					continue
				}
				result = append(result, models.ResponsibleOption{
					Team:           team,
					TeamExternalID: teamID,
					Person:         strings.TrimSpace(person),
					ExternalID:     externalID,
				})
			}
			continue
		}

		externalID := pickStringFromMap(item, "externalId", "ExternalID", "responsibleEmployeeId", "employeeId", "id")
		person := pickStringFromMap(item, "person", "Person", "responsibleEmployeeTitle", "employeeTitle", "title", "name", "fio")
		teamFlat := pickStringFromMap(item, "team", "Team", "responsibleTeamTitle", "teamTitle", "subdivision")
		if teamFlat != "" {
			team = teamFlat
		}
		role := pickStringFromMap(item, "role", "Role", "post", "position")
		if externalID == "" && person == "" {
			continue
		}
		result = append(result, models.ResponsibleOption{
			Team:       team,
			Person:     strings.TrimSpace(person),
			Role:       role,
			ExternalID: externalID,
		})
	}

	return result, nil
}

func parseTicketNumbers(payload []byte) ([]string, error) {
	clean := bytes.TrimPrefix(payload, []byte{0xEF, 0xBB, 0xBF})
	rawText := strings.TrimSpace(string(clean))
	if rawText == "" {
		return []string{}, nil
	}

	// Формат list_sc_responsible: ["0000019683", ...]
	var numberList []string
	if err := json.Unmarshal(clean, &numberList); err == nil {
		return normalizeServiceCallNumbers(numberList), nil
	}

	// Вдруг backend 1С вернёт массив объектов.
	var objectList []map[string]any
	if err := json.Unmarshal(clean, &objectList); err == nil {
		numbers := make([]string, 0, len(objectList))
		for _, item := range objectList {
			number := pickStringFromMap(item, "number", "Number", "sc_number", "sc", "ServiceCallNumber", "Номер")
			if number == "" {
				continue
			}
			numbers = append(numbers, number)
		}
		return normalizeServiceCallNumbers(numbers), nil
	}

	return nil, fmt.Errorf("decode list_sc_responsible response: %s", rawText)
}

// GetTicket returns one detailed ticket card.
func (c *Client) GetTicket(ctx context.Context, userID string, number string) (models.TicketDetail, error) {
	query := map[string]string{
		"id":        strings.TrimSpace(userID),
		"sc_number": strings.TrimSpace(number),
	}

	raw, err := c.doJSONRequestPayload(ctx, http.MethodGet, "/find_sc", query, nil)
	if err != nil {
		return models.TicketDetail{}, err
	}

	payload, err := unmarshalItiliumObjectOrStringError(raw)
	if err != nil {
		return models.TicketDetail{}, err
	}

	detail := parseFindSCResponse(payload, number)
	if strings.TrimSpace(detail.Number) == "" {
		detail.Number = strings.TrimSpace(number)
	}
	if strings.TrimSpace(detail.State) == "" {
		detail.State = "Зарегистрирована"
	}

	return detail, nil
}

// pathCreateSC — HTTP-сервис 1С «создать заявку» (согласовано: id, shortDescription, description, files).
const pathCreateSC = "/create_sc"

// pathCreateSCMarketing — legacy endpoint for marketing requests with dynamic form number.
const pathCreateSCMarketing = "/create_sc_Marketing"

// CreateTicket sends a create request to ITILIUM.
func (c *Client) CreateTicket(ctx context.Context, request models.CreateTicketRequest) (models.TicketDetail, error) {
	// Всегда multipart/form-data (поля + опционально files). Urlencoded на этом HTTP-сервисе 1С ломает разбор запроса.
	return c.doCreateSCMultipart(ctx, request)
}

// ListMarketingServices returns available marketing request types and normalized dynamic form schemas.
func (c *Client) ListMarketingServices(ctx context.Context, userID string) ([]models.MarketingServiceType, error) {
	form := url.Values{}
	form.Set("id", strings.TrimSpace(userID))

	payload, err := c.doFormGetBytes(ctx, "/listServicesMarketing", form)
	if err != nil {
		return nil, err
	}

	services, err := parseMarketingServicesResponse(payload)
	if err != nil {
		return nil, err
	}
	for idx := range services {
		if strings.TrimSpace(services[idx].FormSchema.FormNumber) == "" {
			services[idx].FormSchema.FormNumber = services[idx].FormNumber
		}
		if len(services[idx].FormSchema.Fields) == 0 {
			services[idx].FormSchema = defaultMarketingFormSchema(services[idx].FormNumber, services[idx].Name)
		}
	}

	return services, nil
}

// ListMarketingSubdivisions returns available subdivision options for marketing requests.
func (c *Client) ListMarketingSubdivisions(ctx context.Context, userID string) ([]models.MarketingSubdivision, error) {
	form := url.Values{}
	form.Set("id", strings.TrimSpace(userID))

	payload, err := c.doFormGetBytes(ctx, "/listSubdivisionMarketing", form)
	if err != nil {
		return nil, err
	}

	return parseMarketingSubdivisionsResponse(payload)
}

// CreateMarketingRequest sends dynamic-form marketing payload to legacy create_sc_Marketing endpoint.
func (c *Client) CreateMarketingRequest(ctx context.Context, request models.CreateMarketingRequest) (models.TicketDetail, error) {
	form := buildMarketingCreateForm(request)

	var (
		payload []byte
		err     error
	)
	if len(request.FileAttachments) > 0 {
		payload, err = c.doMultipartFormPostBytesWithFiles(ctx, pathCreateSCMarketing, form, request.FileAttachments)
	} else {
		payload, err = c.submitMarketingCreateWithoutFiles(ctx, form, request.ExecutionDate)
	}
	if err != nil {
		return models.TicketDetail{}, err
	}

	// По поведению legacy-методов ответ может быть пустым или частичным JSON — используем общий parser + fallback.
	createFallback := models.CreateTicketRequest{
		UserID:        request.UserID,
		RequestType:   request.ServiceCode,
		Title:         "Маркетинговая заявка",
		Description:   buildMarketingDescription(request),
		Department:    request.Subdivision,
		ExecutionDate: request.ExecutionDate,
	}
	return parseCreateSCResponse(payload, createFallback)
}

func buildMarketingCreateForm(request models.CreateMarketingRequest) url.Values {
	form := url.Values{}
	form.Set("id", strings.TrimSpace(request.UserID))
	setMarketingCreateFieldAliases(form, []string{"Services", "Service", "services", "КомпонентаУслуги"}, request.ServiceCode)
	setMarketingCreateFieldAliases(form, []string{"Subdivision", "subdivision", "Подразделение"}, request.Subdivision)
	if trimmedDate := strings.TrimSpace(request.ExecutionDate); trimmedDate != "" {
		// В query/body отправляем одно документированное поле; алиасы даты перебираются отдельно в retry.
		// WithoutDate=Ложь вместе с датой на части публикаций 1С приводит к игнорированию ExecutionDate.
		form.Set("ExecutionDate", formatMarketingExecutionDate(trimmedDate))
	} else if request.WithoutDate {
		setMarketingCreateFieldAliases(form, []string{"WithoutDate", "withoutDate", "БезДаты"}, "Истина")
	}
	if formNumber := strings.TrimSpace(request.FormNumber); formNumber != "" {
		setMarketingCreateFieldAliases(form, []string{"FormNumber", "formNumber", "НомерФормы"}, formNumber)
	}
	for key, value := range request.FormData {
		normalizedKey := marketingCreateSCFieldName(key)
		if normalizedKey == "" {
			continue
		}
		form.Set(normalizedKey, strings.TrimSpace(value))
	}
	return form
}

// submitMarketingCreateWithoutFiles отправляет create_sc_Marketing без вложений.
// На текущей публикации 1С query принимает услугу/подразделение; дату пробуем в query, body и multipart по одному полю.
func (c *Client) submitMarketingCreateWithoutFiles(ctx context.Context, form url.Values, rawExecutionDate string) ([]byte, error) {
	query, dateBody := splitMarketingCreateForm(form)

	var lastPayload []byte

	// Минимальный query без дублей алиасов (на тестовом контуре 1С читает услугу/подразделение из query, не из multipart body).
	for idx, postmanQuery := range buildMarketingCreatePostmanQueryVariants(form) {
		payload, err := c.doPostEmptyQuery(ctx, pathCreateSCMarketing, postmanQuery)
		if err == nil && !isMarketingRequiredFieldsMissingResponse(payload) {
			return payload, nil
		}
		if err != nil && !isRetryableMarketingCreateError(err) {
			return nil, err
		}
		if err == nil {
			lastPayload = payload
			if !isMarketingExecutionDateOnlyMissingResponse(payload) {
				break
			}
			continue
		}
		c.logger.Warn(
			"create_sc_Marketing postman query variant failed with retryable error",
			"variant_index", idx,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
	}

	// Сначала — все поля в query, дата в формате YYYY-MM-DD (контракт 1С после фикса).
	if isoDate := formatMarketingExecutionDate(rawExecutionDate); isoDate != "" {
		singleDateQuery := cloneURLValues(query)
		singleDateQuery.Set("ExecutionDate", isoDate)
		payload, err := c.doPostEmptyQuery(ctx, pathCreateSCMarketing, singleDateQuery)
		if err == nil && !isMarketingRequiredFieldsMissingResponse(payload) {
			return payload, nil
		}
		if err != nil && !isRetryableMarketingCreateError(err) {
			return nil, err
		}
		lastPayload = payload
	}

	if len(dateBody) > 0 {
		payload, err := c.doPostQueryWithUrlencodedBody(ctx, pathCreateSCMarketing, query, dateBody)
		if err == nil && !isMarketingRequiredFieldsMissingResponse(payload) {
			return payload, nil
		}
		if err != nil && !isRetryableMarketingCreateError(err) {
			return nil, err
		}
		if err == nil {
			lastPayload = payload
		}
		if err == nil && isMarketingExecutionDateOnlyMissingResponse(payload) {
			for idx, variant := range minimalMarketingDateBodyVariants(rawExecutionDate) {
				variantPayload, variantErr := c.doPostQueryWithUrlencodedBody(ctx, pathCreateSCMarketing, query, variant)
				if variantErr != nil {
					if isRetryableMarketingCreateError(variantErr) {
						c.logger.Warn(
							"create_sc_Marketing date body variant failed with retryable error",
							"variant_index", idx,
							"request_id", middleware.RequestIDFromContext(ctx),
							"user_id", middleware.UserIDFromContext(ctx),
						)
						continue
					}
					return nil, variantErr
				}
				if !isMarketingRequiredFieldsMissingResponse(variantPayload) {
					return variantPayload, nil
				}
				lastPayload = variantPayload
			}
		}
	}

	for idx, variant := range marketingDateFieldVariants(rawExecutionDate) {
		mergedQuery := cloneURLValues(query)
		for key, values := range variant {
			for _, value := range values {
				mergedQuery.Set(key, value)
			}
		}

		payload, err := c.doPostEmptyQuery(ctx, pathCreateSCMarketing, mergedQuery)
		if err == nil && !isMarketingRequiredFieldsMissingResponse(payload) {
			return payload, nil
		}
		if err != nil && !isRetryableMarketingCreateError(err) {
			return nil, err
		}
		if err == nil {
			lastPayload = payload
			if !isMarketingExecutionDateOnlyMissingResponse(payload) {
				break
			}
			continue
		}

		c.logger.Warn(
			"create_sc_Marketing single-date query variant failed with retryable error",
			"variant_index", idx,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
	}

	for idx, variant := range marketingDateFieldVariants(rawExecutionDate) {
		payload, err := c.doPostQueryWithMultipartForm(ctx, pathCreateSCMarketing, query, variant, nil)
		if err == nil && !isMarketingRequiredFieldsMissingResponse(payload) {
			return payload, nil
		}
		if err != nil && !isRetryableMarketingCreateError(err) {
			return nil, err
		}
		if err == nil {
			lastPayload = payload
			if !isMarketingExecutionDateOnlyMissingResponse(payload) {
				break
			}
			continue
		}

		c.logger.Warn(
			"create_sc_Marketing query+multipart date variant failed with retryable error",
			"variant_index", idx,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
	}

	c.logger.Warn(
		"create_sc_Marketing date variants were not accepted, retrying with full date alias set in query",
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)
	fullForm := cloneURLValues(form)
	appendMarketingExecutionDateAliases(fullForm, rawExecutionDate)
	payload, err := c.doPostEmptyQuery(ctx, pathCreateSCMarketing, fullForm)
	if err != nil {
		return nil, err
	}
	if isMarketingRequiredFieldsMissingResponse(payload) && lastPayload != nil {
		return lastPayload, nil
	}
	return payload, nil
}

func isRetryableMarketingCreateError(err error) bool {
	var statusErr HTTPStatusError
	return errors.As(err, &statusErr) && statusErr.StatusCode >= http.StatusInternalServerError
}

func cloneURLValues(source url.Values) url.Values {
	cloned := url.Values{}
	for key, values := range source {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func isMarketingRequiredFieldsMissingResponse(payload []byte) bool {
	message := decodeMarketingResponseMessage(payload)
	if message == "" {
		return false
	}
	lower := strings.ToLower(message)
	return strings.Contains(lower, "не указана услуга") ||
		strings.Contains(lower, "не указано подразделение") ||
		strings.Contains(lower, "желаемую дату исполнения")
}

func isMarketingExecutionDateOnlyMissingResponse(payload []byte) bool {
	message := decodeMarketingResponseMessage(payload)
	if message == "" {
		return false
	}
	lower := strings.ToLower(message)
	return strings.Contains(lower, "желаемую дату исполнения") &&
		!strings.Contains(lower, "не указана услуга") &&
		!strings.Contains(lower, "не указано подразделение")
}

func decodeMarketingResponseMessage(payload []byte) string {
	message := strings.TrimSpace(strings.TrimPrefix(string(bytes.TrimPrefix(payload, []byte{0xEF, 0xBB, 0xBF})), "\ufeff"))
	if message == "" {
		return ""
	}
	var decoded string
	if err := json.Unmarshal([]byte(message), &decoded); err == nil {
		message = strings.TrimSpace(decoded)
	}
	return message
}

func setMarketingCreateFieldAliases(form url.Values, keys []string, value string) {
	trimmed := strings.TrimSpace(value)
	for _, key := range keys {
		form.Set(key, trimmed)
	}
}

func setMarketingExecutionDateAliases(form url.Values, rawDate string) {
	appendMarketingExecutionDateAliases(form, rawDate)
}

func appendMarketingExecutionDateAliases(form url.Values, rawDate string) {
	isoDate := formatMarketingExecutionDate(rawDate)
	if isoDate == "" {
		return
	}

	calendarDate := formatItiliumCalendarDate(rawDate)
	calendarDateTime := formatItiliumCalendarDateTime(rawDate)

	// 1С на разных публикациях ожидает ISO, календарную дату или дату-время как в find_sc.
	for key, value := range map[string]string{
		"ExecutionDate":          calendarDateTime,
		"ЖелаемаяДатаИсполнения": calendarDateTime,
		"ДатаИсполнения":         calendarDate,
		"executionDate":          isoDate,
		"DesiredExecutionDate":   isoDate,
		"ExecutionDateISO":       isoDate,
		"date_inc":               isoDate,
	} {
		if value != "" {
			form.Set(key, value)
		}
	}
}

func marketingExecutionDateFieldKeys() map[string]struct{} {
	return map[string]struct{}{
		"ExecutionDate":          {},
		"executionDate":          {},
		"ДатаИсполнения":         {},
		"ЖелаемаяДатаИсполнения": {},
		"DesiredExecutionDate":   {},
		"deadlineDate":           {},
		"Date":                   {},
		"ExecutionDateTime":      {},
		"ExecutionDateISO":       {},
		"date_inc":               {},
		"СрокИсполнения":         {},
		"ДатаВыполнения":         {},
		"ЖелаемаяДата":           {},
	}
}

func marketingDateFieldVariants(rawDate string) []url.Values {
	isoDate := formatMarketingExecutionDate(rawDate)
	if isoDate == "" {
		return nil
	}
	calendarDate := formatItiliumCalendarDate(rawDate)
	calendarDateTime := formatItiliumCalendarDateTime(rawDate)

	candidates := []struct {
		key   string
		value string
	}{
		{"ExecutionDate", calendarDateTime},
		{"ЖелаемаяДатаИсполнения", calendarDateTime},
		{"ExecutionDate", calendarDate},
		{"ЖелаемаяДатаИсполнения", calendarDate},
		{"ExecutionDate", isoDate},
		{"ЖелаемаяДатаИсполнения", isoDate},
		{"ДатаИсполнения", calendarDateTime},
		{"ДатаИсполнения", isoDate},
		{"DesiredExecutionDate", isoDate},
		{"date_inc", isoDate},
	}

	variants := make([]url.Values, 0, len(candidates))
	for _, candidate := range candidates {
		variants = append(variants, url.Values{candidate.key: {candidate.value}})
	}
	return variants
}

// marketingCreateMultipartForm готовит поля для POST multipart как в Postman (без WithoutDate при указанной дате).
func marketingCreateMultipartForm(form url.Values) url.Values {
	multipartForm := cloneURLValues(form)
	if formatMarketingExecutionDate(firstMarketingExecutionDate(multipartForm)) != "" {
		for _, key := range []string{"WithoutDate", "withoutDate", "БезДаты"} {
			multipartForm.Del(key)
		}
	}
	return multipartForm
}

func firstMarketingExecutionDate(form url.Values) string {
	dateKeys := marketingExecutionDateFieldKeys()
	for key, values := range form {
		if _, isDate := dateKeys[key]; !isDate {
			continue
		}
		for _, value := range values {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func splitMarketingCreateForm(form url.Values) (url.Values, url.Values) {
	dateKeys := marketingExecutionDateFieldKeys()
	query := url.Values{}
	dateBody := url.Values{}
	for key, values := range form {
		if _, isDateField := dateKeys[key]; isDateField {
			for _, value := range values {
				dateBody.Add(key, value)
			}
			continue
		}
		for _, value := range values {
			query.Add(key, value)
		}
	}
	return query, dateBody
}

func minimalMarketingDateBodyVariants(rawDate string) []url.Values {
	isoDate := formatMarketingExecutionDate(rawDate)
	if isoDate == "" {
		return nil
	}
	calendarDateTime := formatItiliumCalendarDateTime(rawDate)

	return []url.Values{
		{"ExecutionDate": {calendarDateTime}},
		{"ЖелаемаяДатаИсполнения": {calendarDateTime}},
		{"ExecutionDate": {isoDate}},
		{"ЖелаемаяДатаИсполнения": {isoDate}},
		{"ДатаИсполнения": {isoDate}},
	}
}

// decodeItiliumResponseMessage извлекает текстовое сообщение из JSON-строки 1С (HTTP 200).
func decodeItiliumResponseMessage(payload []byte) string {
	return decodeMarketingResponseMessage(payload)
}

// isItiliumMutationSuccessMessage распознаёт подтверждения 1С при HTTP 200 (не путать с ошибками).
func isItiliumMutationSuccessMessage(message string) bool {
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		return true
	}
	successPhrases := []string{
		"новое состояние установлено",
		"состояние установлено",
		"ответственный изменен",
		"ответственный назначен",
		"успешно",
	}
	for _, phrase := range successPhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

// isItiliumMutationBusinessErrorMessage распознаёт текстовые ошибки 1С в теле при HTTP 200.
func isItiliumMutationBusinessErrorMessage(message string) bool {
	if isItiliumMutationSuccessMessage(message) {
		return false
	}
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		return false
	}
	errorSubstrings := []string{
		"не заполнен",
		"ошибк",
		"не может",
		"преобразование",
		"неверн",
		"отказ",
		"не удалось",
		"недопустим",
		"запрещ",
		"не указан",
		"не указано",
		"не найден",
	}
	for _, sub := range errorSubstrings {
		if strings.Contains(normalized, sub) {
			return true
		}
	}
	// Неизвестное непустое сообщение — считаем ошибкой, чтобы не пропустить новые отказы 1С.
	return true
}

// parseItiliumMutationResponse трактует тело multipart-мутации 1С при HTTP 200.
func parseItiliumMutationResponse(payload []byte) error {
	message := decodeItiliumResponseMessage(payload)
	if !isItiliumMutationBusinessErrorMessage(message) {
		return nil
	}
	return errors.New(message)
}

// formatItiliumCalendarDate приводит дату к формату DD.MM.YYYY, как в полях creationDate/deadlineDate 1С.
func formatItiliumCalendarDate(value string) string {
	isoDate := formatMarketingExecutionDate(value)
	if isoDate == "" {
		return ""
	}
	parsed, err := time.Parse("2006-01-02", isoDate)
	if err != nil {
		return ""
	}
	return parsed.Format("02.01.2006")
}

// formatItiliumCalendarDateTime приводит дату к DD.MM.YYYY H:MM:SS — как creationDate в find_sc.
func formatItiliumCalendarDateTime(value string) string {
	calendarDate := formatItiliumCalendarDate(value)
	if calendarDate == "" {
		return ""
	}
	return calendarDate + " 0:00:00"
}

// buildMarketingCreatePostmanQueryVariants возвращает минимальные query с датой в разных форматах 1С.
func buildMarketingCreatePostmanQueryVariants(form url.Values) []url.Values {
	rawDate := firstMarketingExecutionDate(form)
	dateFormats := []string{
		formatMarketingExecutionDate(rawDate),
		formatItiliumCalendarDate(rawDate),
		formatItiliumCalendarDateTime(rawDate),
	}
	seen := make(map[string]struct{}, len(dateFormats))
	variants := make([]url.Values, 0, len(dateFormats))
	for _, dateValue := range dateFormats {
		if dateValue == "" {
			continue
		}
		if _, exists := seen[dateValue]; exists {
			continue
		}
		seen[dateValue] = struct{}{}
		variants = append(variants, buildMarketingCreatePostmanQuery(form, dateValue))
	}
	return variants
}

// buildMarketingCreatePostmanQuery собирает минимальный набор полей create_sc_Marketing (без дублей алиасов).
func buildMarketingCreatePostmanQuery(form url.Values, executionDate string) url.Values {
	query := url.Values{}
	if id := strings.TrimSpace(form.Get("id")); id != "" {
		query.Set("id", id)
	}
	if service := firstNonEmptyFormValue(form, "Services", "Service", "services", "КомпонентаУслуги"); service != "" {
		query.Set("Services", service)
	}
	if subdivision := firstNonEmptyFormValue(form, "Subdivision", "subdivision", "Подразделение"); subdivision != "" {
		query.Set("Subdivision", subdivision)
	}
	if formNumber := firstNonEmptyFormValue(form, "FormNumber", "formNumber", "НомерФормы"); formNumber != "" {
		query.Set("FormNumber", formNumber)
	}
	if description := strings.TrimSpace(form.Get("Description")); description != "" {
		query.Set("Description", description)
	}
	if trimmedDate := strings.TrimSpace(executionDate); trimmedDate != "" {
		query.Set("ExecutionDate", trimmedDate)
	}
	for key, values := range form {
		if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
			continue
		}
		if _, skip := marketingExecutionDateFieldKeys()[key]; skip {
			continue
		}
		switch key {
		case "id", "Services", "Service", "services", "КомпонентаУслуги",
			"Subdivision", "subdivision", "Подразделение",
			"FormNumber", "formNumber", "НомерФормы", "Description",
			"WithoutDate", "withoutDate", "БезДаты":
			continue
		}
		query.Set(key, strings.TrimSpace(values[0]))
	}
	return query
}

func firstNonEmptyFormValue(form url.Values, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(form.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

// buildChangeStateForm собирает поля для POST /change_state_sc по контракту 1С.
func buildChangeStateForm(number string, request models.ChangeStatusRequest) url.Values {
	form := url.Values{}
	userID := strings.TrimSpace(request.UserID)
	trimmedNumber := strings.TrimSpace(number)
	// Совместимость с legacy-обработчиком: передаём оба идентификатора пользователя.
	form.Set("id", userID)
	form.Set("telegram", userID)
	form.Set("inc_number", trimmedNumber)
	form.Set("new_state", strings.TrimSpace(request.State))
	// Для «Отложено» и других переходов с комментарием 1С ждёт comment_text (не comment).
	if comment := strings.TrimSpace(request.Comment); comment != "" {
		form.Set("comment_text", comment)
	}
	if rawDate := strings.TrimSpace(request.Date); rawDate != "" {
		if date := formatChangeStateDateInc(rawDate); date != "" {
			form.Set("date_inc", date)
		}
	}
	return form
}

// formatChangeStateDateInc нормализует date_inc для change_state_sc (формат YYYY-MM-DD в 1С).
func formatChangeStateDateInc(rawDate string) string {
	return formatMarketingExecutionDate(rawDate)
}

// formatMarketingExecutionDate нормализует дату для create_sc_Marketing в формат YYYY-MM-DD.
func formatMarketingExecutionDate(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	if parsed, err := time.Parse("2006-01-02", trimmed); err == nil {
		return parsed.Format("2006-01-02")
	}
	if parsed, err := time.Parse("02.01.2006", trimmed); err == nil {
		return parsed.Format("2006-01-02")
	}

	return trimmed
}

func marketingCreateSCFieldName(key string) string {
	switch strings.TrimSpace(key) {
	case "LayoutName", "layoutName":
		return "LayoutName"
	case "Size", "size":
		return "Size"
	case "ForWhat", "purpose", "forWhat":
		return "ForWhat"
	case "RequiredText", "requiredText":
		return "RequiredText"
	case "LayoutFormats", "layoutFormat", "layoutFormats":
		return "LayoutFormats"
	case "ThemeEvent", "eventTheme":
		return "ThemeEvent"
	case "Description", "eventDescription", "details":
		return "Description"
	case "Budget", "budget":
		return "Budget"
	case "LinkToFoto", "photoLinks", "photoLink":
		return "LinkToFoto"
	case "LinkToExamples", "referenceLinks", "referenceLink":
		return "LinkToExamples"
	default:
		return strings.TrimSpace(key)
	}
}

func buildCreateSCLongDescription(req models.CreateTicketRequest) string {
	// Обычный create_sc ждёт в description только пользовательский текст.
	// Тип, подразделение и срок остаются отдельной логикой маркетингового endpoint и не подмешиваются сюда.
	return strings.TrimSpace(req.Description)
}

func parseCreateSCResponse(payload []byte, req models.CreateTicketRequest) (models.TicketDetail, error) {
	payload = bytes.TrimPrefix(payload, []byte{0xEF, 0xBB, 0xBF})
	if len(strings.TrimSpace(string(payload))) == 0 {
		return ticketDetailCreateSCFallback(req), nil
	}

	var message string
	if err := json.Unmarshal(payload, &message); err == nil {
		message = strings.TrimSpace(message)
		if message == "" {
			return ticketDetailCreateSCFallback(req), nil
		}
		return models.TicketDetail{}, errors.New(message)
	}

	var detail models.TicketDetail
	if err := json.Unmarshal(payload, &detail); err == nil && strings.TrimSpace(detail.Number+detail.Title) != "" {
		if detail.Title == "" {
			detail.Title = req.Title
		}
		if detail.Description == "" {
			detail.Description = req.Description
		}
		if detail.State == "" {
			detail.State = "Зарегистрирована"
		}
		return detail, nil
	}

	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return models.TicketDetail{}, fmt.Errorf("decode create_sc response: %w", err)
	}

	if len(m) == 0 {
		return ticketDetailCreateSCFallback(req), nil
	}

	detail = models.TicketDetail{
		Number:               pickStringFromMap(m, "number", "Number", "sc", "Номер", "ServiceCallNumber"),
		Title:                pickStringFromMap(m, "title", "Title", "shortDescription", "ShortDescription"),
		Description:          pickStringFromMap(m, "description", "Description"),
		State:                pickStringFromMap(m, "state", "State", "status", "Status"),
		CreationDate:         pickStringFromMap(m, "creationDate", "CreationDate", "dateCreate", "ДатаСоздания"),
		Deadline:             pickStringFromMap(m, "deadline", "Deadline", "deadlineDate", "executionDate", "ДатаИсполнения"),
		ResponsibleTeam:      pickStringFromMap(m, "responsibleTeam", "ResponsibleTeam", "OU", "client"),
		CanChangeResponsible: boolFromAny(m["canChangeResponsible"]),
	}
	if detail.Title == "" {
		detail.Title = req.Title
	}
	if detail.Description == "" {
		detail.Description = req.Description
	}
	if detail.State == "" {
		detail.State = "Зарегистрирована"
	}
	if detail.Deadline == "" {
		detail.Deadline = req.ExecutionDate
	}
	return detail, nil
}

func parseMarketingServicesResponse(payload []byte) ([]models.MarketingServiceType, error) {
	clean := bytes.TrimPrefix(payload, []byte{0xEF, 0xBB, 0xBF})
	if len(strings.TrimSpace(string(clean))) == 0 {
		return []models.MarketingServiceType{}, nil
	}

	var list []map[string]any
	if err := json.Unmarshal(clean, &list); err == nil {
		return mapMarketingServices(list), nil
	}

	var envelope map[string]any
	if err := json.Unmarshal(clean, &envelope); err != nil {
		return nil, fmt.Errorf("decode listServicesMarketing response: %w", err)
	}

	for _, key := range []string{"data", "items", "result", "services"} {
		raw, ok := envelope[key]
		if !ok {
			continue
		}
		items, ok := raw.([]any)
		if !ok {
			continue
		}
		normalized := make([]map[string]any, 0, len(items))
		for _, item := range items {
			if m, castOK := item.(map[string]any); castOK {
				normalized = append(normalized, m)
			}
		}
		return mapMarketingServices(normalized), nil
	}

	return []models.MarketingServiceType{}, nil
}

func mapMarketingServices(items []map[string]any) []models.MarketingServiceType {
	out := make([]models.MarketingServiceType, 0, len(items))
	for _, item := range items {
		code := pickStringFromMap(item, "code", "serviceCode", "service", "Services", "Service", "КомпонентаУслуги", "Услуга", "id", "key")
		name := pickStringFromMap(item, "name", "title", "serviceName", "Services", "Service", "КомпонентаУслуги", "Услуга", "label")
		formNumber := pickStringFromMap(item, "formNumber", "FormNumber", "form_number", "form", "numberForm", "NumberForm", "НомерФормы")
		if code == "" && name == "" {
			continue
		}
		if code == "" {
			code = name
		}
		if name == "" {
			name = code
		}
		schema := defaultMarketingFormSchema(formNumber, name)
		out = append(out, models.MarketingServiceType{
			Code:       strings.TrimSpace(code),
			Name:       strings.TrimSpace(name),
			FormNumber: strings.TrimSpace(formNumber),
			FormSchema: schema,
		})
	}
	return out
}

func parseMarketingSubdivisionsResponse(payload []byte) ([]models.MarketingSubdivision, error) {
	clean := bytes.TrimPrefix(payload, []byte{0xEF, 0xBB, 0xBF})
	if len(strings.TrimSpace(string(clean))) == 0 {
		return []models.MarketingSubdivision{}, nil
	}

	var names []string
	if err := json.Unmarshal(clean, &names); err == nil {
		result := make([]models.MarketingSubdivision, 0, len(names))
		for _, name := range names {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				continue
			}
			result = append(result, models.MarketingSubdivision{Name: trimmed})
		}
		return result, nil
	}

	var list []map[string]any
	if err := json.Unmarshal(clean, &list); err == nil {
		result := make([]models.MarketingSubdivision, 0, len(list))
		for _, item := range list {
			name := pickStringFromMap(item, "name", "title", "subdivision", "Subdivision")
			if name == "" {
				continue
			}
			result = append(result, models.MarketingSubdivision{
				Code: pickStringFromMap(item, "code", "id", "subdivisionCode"),
				Name: name,
			})
		}
		return result, nil
	}

	return nil, fmt.Errorf("decode listSubdivisionMarketing response: unsupported format")
}

func buildMarketingDescription(req models.CreateMarketingRequest) string {
	var b strings.Builder
	b.WriteString("Маркетинговая заявка")
	b.WriteString("\nТип: ")
	b.WriteString(strings.TrimSpace(req.ServiceCode))
	b.WriteString("\nПодразделение: ")
	b.WriteString(strings.TrimSpace(req.Subdivision))
	if strings.TrimSpace(req.ExecutionDate) != "" {
		b.WriteString("\nИсполнить до: ")
		b.WriteString(strings.TrimSpace(req.ExecutionDate))
	} else if req.WithoutDate {
		b.WriteString("\nИсполнить до: без даты")
	}
	for key, value := range req.FormData {
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(key))
		b.WriteString(": ")
		b.WriteString(strings.TrimSpace(value))
	}
	return b.String()
}

func defaultMarketingFormSchema(formNumber string, serviceName string) models.MarketingFormSchema {
	normalizedFormNumber := normalizeMarketingFormNumber(formNumber)
	schema := models.MarketingFormSchema{
		FormNumber: normalizedFormNumber,
		Title:      strings.TrimSpace(serviceName),
	}
	// Номер формы приходит из 1С и точнее определяет набор обязательных полей, чем человекочитаемое название услуги.
	switch normalizedFormNumber {
	case "1":
		schema.Fields = []models.MarketingFormField{
			{Key: "LayoutName", Label: "Название макета (баннер, афиша)", Type: "text", Required: true},
			{Key: "Size", Label: "Размеры (в мм или dpi)", Type: "text", Required: true},
			{Key: "ForWhat", Label: "Для чего (печать, веб версия)", Type: "text", Required: true},
			{Key: "RequiredText", Label: "Обязательный текст", Type: "textarea", Required: true},
			{Key: "LayoutFormats", Label: "Форматы представления макета (pdf, png, psd, tiff, crd)", Type: "text", Required: true},
			{Key: "LinkToFoto", Label: "Ссылка на обязательные фотоматериалы (при наличии)", Type: "text", Required: false},
			{Key: "LinkToExamples", Label: "Ссылки на примеры, референсы", Type: "text", Required: false},
		}
	case "2":
		schema.Fields = []models.MarketingFormField{
			{Key: "ThemeEvent", Label: "Тема мероприятия", Type: "text", Required: true},
			{Key: "Description", Label: "Описание", Type: "textarea", Required: true},
			{Key: "Budget", Label: "Бюджет", Type: "text", Required: true},
			{Key: "FreeText", Label: "Свободное текстовое поле", Type: "textarea", Required: false},
		}
	case "3":
		schema.Fields = []models.MarketingFormField{
			{Key: "Description", Label: "Свободное текстовое поле", Type: "textarea", Required: true},
		}
	default:
		switch strings.ToLower(strings.TrimSpace(serviceName)) {
		case "дизайн":
			schema.Fields = []models.MarketingFormField{
				{Key: "LayoutName", Label: "Название макета (баннер, афиша)", Type: "text", Required: true},
				{Key: "Size", Label: "Размеры (в мм или dpi)", Type: "text", Required: true},
				{Key: "ForWhat", Label: "Для чего (печать, веб версия)", Type: "text", Required: true},
				{Key: "RequiredText", Label: "Обязательный текст", Type: "textarea", Required: true},
				{Key: "LayoutFormats", Label: "Форматы представления макета (pdf, png, psd, tiff, crd)", Type: "text", Required: true},
				{Key: "LinkToFoto", Label: "Ссылка на обязательные фотоматериалы (при наличии)", Type: "text", Required: false},
				{Key: "LinkToExamples", Label: "Ссылки на примеры, референсы", Type: "text", Required: false},
			}
		case "мероприятие":
			schema.Fields = []models.MarketingFormField{
				{Key: "ThemeEvent", Label: "Тема мероприятия", Type: "text", Required: true},
				{Key: "Description", Label: "Описание", Type: "textarea", Required: true},
				{Key: "Budget", Label: "Бюджет", Type: "text", Required: true},
				{Key: "FreeText", Label: "Свободное текстовое поле", Type: "textarea", Required: false},
			}
		default:
			schema.Fields = []models.MarketingFormField{
				{Key: "Description", Label: "Свободное текстовое поле", Type: "textarea", Required: true},
			}
		}
	}
	return schema
}

func normalizeMarketingFormNumber(value string) string {
	normalized := strings.TrimSpace(strings.TrimSuffix(value, ".0"))
	if normalized == "" {
		return ""
	}
	withoutLeadingZeros := strings.TrimLeft(normalized, "0")
	if withoutLeadingZeros == "" {
		return "0"
	}
	return withoutLeadingZeros
}

func ticketDetailCreateSCFallback(req models.CreateTicketRequest) models.TicketDetail {
	return models.TicketDetail{
		Title:                req.Title,
		Description:          req.Description,
		State:                "Зарегистрирована",
		Deadline:             req.ExecutionDate,
		ResponsibleTeam:      req.Department,
		CanChangeResponsible: true,
		Timeline: []models.CommentEntry{
			{Author: "Система", Message: "Заявка передана в ITILIUM (ответ create_sc без тела или пустой).", CreatedAt: time.Now().Format(time.RFC3339)},
		},
	}
}

func pickStringFromMap(m map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		s := strings.TrimSpace(fmt.Sprint(v))
		if s != "" {
			return s
		}
	}
	return ""
}

// doCreateSCMultipart отправляет create_sc с полями id, shortDescription, description и файлами в частях files.
func (c *Client) doCreateSCMultipart(ctx context.Context, request models.CreateTicketRequest) (models.TicketDetail, error) {
	if c.baseURL == "" {
		return models.TicketDetail{}, errors.New("itilium base url is required")
	}

	body := &bytes.Buffer{}
	mp := multipart.NewWriter(body)

	if err := mp.WriteField("id", strings.TrimSpace(request.UserID)); err != nil {
		return models.TicketDetail{}, fmt.Errorf("write id: %w", err)
	}
	if err := mp.WriteField("shortDescription", strings.TrimSpace(request.Title)); err != nil {
		return models.TicketDetail{}, fmt.Errorf("write shortDescription: %w", err)
	}
	if err := mp.WriteField("description", buildCreateSCLongDescription(request)); err != nil {
		return models.TicketDetail{}, fmt.Errorf("write description: %w", err)
	}

	for _, fa := range request.FileAttachments {
		part, err := mp.CreateFormFile("files", fa.Filename)
		if err != nil {
			return models.TicketDetail{}, fmt.Errorf("create form file: %w", err)
		}
		if _, err := part.Write(fa.Data); err != nil {
			return models.TicketDetail{}, fmt.Errorf("write file: %w", err)
		}
	}

	if err := mp.Close(); err != nil {
		return models.TicketDetail{}, fmt.Errorf("close multipart writer: %w", err)
	}

	endpoint, err := url.Parse(c.baseURL + pathCreateSC)
	if err != nil {
		return models.TicketDetail{}, fmt.Errorf("parse itilium url: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), body)
	if err != nil {
		return models.TicketDetail{}, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", mp.FormDataContentType())
	if c.login != "" {
		httpReq.SetBasicAuth(c.login, c.password)
	}

	// В INFO пишем то, что уходит в 1С (тело multipart с бинарниками в лог не дублируем — только поля и метаданные файлов).
	longDesc := buildCreateSCLongDescription(request)
	fileMeta := make([]string, 0, len(request.FileAttachments))
	for _, fa := range request.FileAttachments {
		fileMeta = append(fileMeta, fmt.Sprintf("%s:%dB", fa.Filename, len(fa.Data)))
	}
	// Имена ключей ниже — только для slog (не уходят в 1С). В HTTP в 1С поля называются id, shortDescription, description, files.
	c.logger.Info(
		"itilium outbound create_sc (multipart fields)",
		"url", endpoint.String(),
		"send_id", strings.TrimSpace(request.UserID),
		"send_shortDescription", strings.TrimSpace(request.Title),
		"send_description", truncateLogString(longDesc, 6000),
		"send_files_meta", strings.Join(fileMeta, ","),
	)

	start := time.Now()
	response, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error("itilium request failed", "method", http.MethodPost, "url", endpoint.String(), "duration_ms", time.Since(start).Milliseconds(), "error", err)
		return models.TicketDetail{}, fmt.Errorf("perform request: %w", err)
	}
	defer response.Body.Close()

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return models.TicketDetail{}, fmt.Errorf("read response: %w", err)
	}

	respText := strings.TrimPrefix(string(respBody), "\ufeff")

	c.logger.Info(
		"itilium create_sc multipart completed",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"response_body", truncateLogString(respText, 8000),
	)
	c.logger.Debug(
		"itilium create_sc multipart details",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"response_body", respText,
	)

	if response.StatusCode >= 400 {
		return models.TicketDetail{}, HTTPStatusError{StatusCode: response.StatusCode}
	}

	detail, err := parseCreateSCResponse(respBody, request)
	if err != nil {
		return models.TicketDetail{}, err
	}

	// 1С часто отвечает 200 с телом "" — номер заявки подтягиваем из find_employee + find_sc.
	if strings.TrimSpace(detail.Number) == "" {
		resolved, resolveErr := c.resolveTicketDetailAfterCreate(ctx, request)
		if resolveErr == nil {
			c.logger.Info(
				"create_sc returned empty body, resolved ticket via find_employee/find_sc",
				"number", resolved.Number,
				"request_id", middleware.RequestIDFromContext(ctx),
				"user_id", middleware.UserIDFromContext(ctx),
			)
			return resolved, nil
		}
		c.logger.Warn(
			"create_sc returned empty body, could not resolve ticket number",
			"error", resolveErr,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
	}

	return detail, nil
}

// resolveTicketDetailAfterCreate ищет только что созданную заявку по свежему списку servicecalls и find_sc.
func (c *Client) resolveTicketDetailAfterCreate(ctx context.Context, request models.CreateTicketRequest) (models.TicketDetail, error) {
	lookup, err := c.FindEmployeeByIdentifier(ctx, models.EmployeeLookupRequest{
		Identifier:    strings.TrimSpace(request.UserID),
		AttributeCode: "id",
	})
	if err != nil {
		return models.TicketDetail{}, err
	}

	numbers := normalizeServiceCallNumbers(lookup.ServiceCalls)
	if len(numbers) == 0 {
		return models.TicketDetail{}, errors.New("no servicecalls after create")
	}

	wantTitle := strings.TrimSpace(request.Title)
	wantDescription := strings.TrimSpace(request.Description)

	// Свежие номера в начале servicecalls — достаточно нескольких find_sc, не обходим весь список.
	const resolveAfterCreateMaxLookups = 5
	lookupLimit := len(numbers)
	if lookupLimit > resolveAfterCreateMaxLookups {
		lookupLimit = resolveAfterCreateMaxLookups
	}

	for _, number := range numbers[:lookupLimit] {
		detail, err := c.GetTicket(ctx, request.UserID, number)
		if err != nil {
			continue
		}
		if wantTitle != "" && strings.TrimSpace(detail.Title) == wantTitle {
			return detail, nil
		}
		if wantDescription != "" && strings.TrimSpace(detail.Description) == wantDescription {
			return detail, nil
		}
	}

	return c.GetTicket(ctx, request.UserID, numbers[0])
}

// AddComment adds a new comment to a ticket.
func (c *Client) AddComment(ctx context.Context, number string, request models.AddCommentRequest) (models.TicketDetail, error) {
	commentText := strings.TrimSpace(request.Message)
	// Если текст пустой, но есть файлы, подставляем короткую подпись — поле comment_text в 1С не должно оставаться пустым.
	if commentText == "" && len(request.FileAttachments) > 0 {
		commentText = "Прикреплён файл."
	}

	var err error
	if len(request.FileAttachments) > 0 {
		// Вложения уходят в те же части files, что и при create_sc (контракт HTTP-сервиса 1С).
		err = c.doAddCommentMultipart(ctx, number, commentText, request)
	} else {
		form := url.Values{}
		form.Set("id", strings.TrimSpace(request.UserID))
		form.Set("source", strings.TrimSpace(number))
		form.Set("source_type", "servicecall")
		form.Set("comment_text", commentText)
		_, err = c.doMultipartFormPostBytes(ctx, "/add_comment", form)
	}
	if err != nil {
		return models.TicketDetail{}, err
	}

	// Возвращаем актуальную карточку после добавления комментария.
	detail, err := c.GetTicket(ctx, request.UserID, number)
	if err != nil {
		// На случай временной недоступности find_sc возвращаем минимально обновлённую карточку.
		fallbackMsg := strings.TrimSpace(request.Message)
		if fallbackMsg == "" && len(request.FileAttachments) > 0 {
			fallbackMsg = "Прикреплён файл."
		}
		return models.TicketDetail{
			Number: number,
			Timeline: []models.CommentEntry{
				{
					Author:    "Пользователь",
					Message:   fallbackMsg,
					CreatedAt: time.Now().Format(time.RFC3339),
				},
			},
		}, nil
	}

	return detail, nil
}

// doAddCommentMultipart отправляет add_comment с полями формы и файлами в частях files.
func (c *Client) doAddCommentMultipart(ctx context.Context, scNumber string, commentText string, request models.AddCommentRequest) error {
	if c.baseURL == "" {
		return errors.New("itilium base url is required")
	}

	body := &bytes.Buffer{}
	mp := multipart.NewWriter(body)

	if err := mp.WriteField("id", strings.TrimSpace(request.UserID)); err != nil {
		return fmt.Errorf("write id: %w", err)
	}
	if err := mp.WriteField("source", strings.TrimSpace(scNumber)); err != nil {
		return fmt.Errorf("write source: %w", err)
	}
	if err := mp.WriteField("source_type", "servicecall"); err != nil {
		return fmt.Errorf("write source_type: %w", err)
	}
	if err := mp.WriteField("comment_text", commentText); err != nil {
		return fmt.Errorf("write comment_text: %w", err)
	}

	for _, fa := range request.FileAttachments {
		part, err := mp.CreateFormFile("files", fa.Filename)
		if err != nil {
			return fmt.Errorf("create form file: %w", err)
		}
		if _, err := part.Write(fa.Data); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
	}

	if err := mp.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	endpoint, err := url.Parse(c.baseURL + "/add_comment")
	if err != nil {
		return fmt.Errorf("parse itilium url: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", mp.FormDataContentType())
	if c.login != "" {
		httpReq.SetBasicAuth(c.login, c.password)
	}

	fileMeta := make([]string, 0, len(request.FileAttachments))
	for _, fa := range request.FileAttachments {
		fileMeta = append(fileMeta, fmt.Sprintf("%s:%dB", fa.Filename, len(fa.Data)))
	}
	c.logger.Info(
		"itilium outbound add_comment (multipart fields)",
		"url", endpoint.String(),
		"send_id", strings.TrimSpace(request.UserID),
		"send_source", strings.TrimSpace(scNumber),
		"send_comment_text", truncateLogString(commentText, 4000),
		"send_files_meta", strings.Join(fileMeta, ","),
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)

	start := time.Now()
	response, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error(
			"itilium add_comment multipart failed",
			"url", endpoint.String(),
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
		return fmt.Errorf("perform request: %w", err)
	}
	defer response.Body.Close()

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	respText := strings.TrimPrefix(string(respBody), "\ufeff")
	c.logger.Info(
		"itilium add_comment multipart completed",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"response_body", truncateLogString(respText, 8000),
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)
	c.logger.Debug(
		"itilium add_comment multipart details",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"response_body", respText,
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)

	if response.StatusCode >= 400 {
		return HTTPStatusError{StatusCode: response.StatusCode}
	}

	return nil
}

// ChangeStatus changes the workflow status of a ticket.
func (c *Client) ChangeStatus(ctx context.Context, number string, request models.ChangeStatusRequest) (models.TicketDetail, error) {
	form := buildChangeStateForm(number, request)

	payload, err := c.doMultipartFormPostBytes(ctx, "/change_state_sc", form)
	if err != nil {
		return models.TicketDetail{}, err
	}
	// 1С может вернуть HTTP 200 с JSON-строкой ошибки вместо пустого тела.
	if err := parseItiliumMutationResponse(payload); err != nil {
		return models.TicketDetail{}, err
	}

	// После смены статуса возвращаем обновлённую карточку из find_sc.
	detail, err := c.GetTicket(ctx, request.UserID, number)
	if err != nil {
		return models.TicketDetail{
			Number: number,
			State:  request.State,
		}, nil
	}
	requestedState := strings.TrimSpace(request.State)
	if requestedState != "" && strings.TrimSpace(detail.State) != requestedState {
		// Успешный change_state_sc, но find_sc ещё отдаёт старую карточку — подменяем только после подтверждённого вызова 1С.
		c.logger.Warn(
			"find_sc returned stale state after change_state_sc",
			"number", number,
			"old_state", detail.State,
			"new_state", requestedState,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
		detail.State = requestedState
	}
	return detail, nil
}

// buildChangeResponsibleForm собирает поля POST /change_responsible_sc по контракту 1С.
func buildChangeResponsibleForm(number string, request models.ChangeResponsibleRequest) url.Values {
	form := url.Values{}
	userID := strings.TrimSpace(request.UserID)
	trimmedNumber := strings.TrimSpace(number)
	form.Set("id", userID)
	form.Set("telegram", userID)
	form.Set("inc_number", trimmedNumber)
	// Дублируем номер: часть обработчиков 1С читает sc_number, а не inc_number.
	form.Set("sc_number", trimmedNumber)
	form.Set("responsibleEmployeeId", strings.TrimSpace(request.ResponsibleID))
	if teamID := strings.TrimSpace(request.TeamID); teamID != "" {
		form.Set("responsibleTeamId", teamID)
	}
	return form
}

// responsibleChangePollDelays — паузы перед повторным find_sc: 1С часто отдаёт старого исполнителя сразу после 204.
var responsibleChangePollDelays = []time.Duration{
	0,
	300 * time.Millisecond,
	800 * time.Millisecond,
	1500 * time.Millisecond,
	3 * time.Second,
	5 * time.Second,
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// pollTicketResponsibleAfterChange ждёт, пока find_sc вернёт выбранного исполнителя, либо отдаёт последний ответ.
func (c *Client) pollTicketResponsibleAfterChange(ctx context.Context, userID, number, requestedID string) (models.TicketDetail, error) {
	trimmedID := strings.TrimSpace(requestedID)
	var detail models.TicketDetail
	var lastErr error

	for _, delay := range responsibleChangePollDelays {
		if err := sleepContext(ctx, delay); err != nil {
			return detail, err
		}
		detail, lastErr = c.GetTicket(ctx, userID, number)
		if lastErr != nil {
			continue
		}
		if trimmedID == "" || strings.TrimSpace(detail.ResponsibleEmployeeID) == trimmedID {
			return detail, nil
		}
	}

	if detail.Number != "" {
		return detail, nil
	}
	if lastErr != nil {
		return models.TicketDetail{Number: number}, lastErr
	}
	return models.TicketDetail{Number: number}, nil
}

// resolveResponsibleTeamID подставляет teamId из responsibles_sc, если UI его не передал.
func (c *Client) resolveResponsibleTeamID(
	ctx context.Context,
	userID, number string,
	request models.ChangeResponsibleRequest,
) models.ChangeResponsibleRequest {
	if strings.TrimSpace(request.TeamID) != "" {
		return request
	}
	options, err := c.ListResponsibleOptions(ctx, userID, number)
	if err != nil {
		return request
	}
	if option, ok := responsibleOptionMatch(options, request.ResponsibleID, ""); ok && strings.TrimSpace(option.TeamExternalID) != "" {
		request.TeamID = option.TeamExternalID
	}
	return request
}

// ChangeResponsible changes the responsible person of a ticket.
func (c *Client) ChangeResponsible(ctx context.Context, number string, request models.ChangeResponsibleRequest) (models.TicketDetail, error) {
	request = c.resolveResponsibleTeamID(ctx, request.UserID, number, request)
	form := buildChangeResponsibleForm(number, request)

	payload, err := c.doMultipartFormPostBytes(ctx, "/change_responsible_sc", form)
	if err != nil {
		return models.TicketDetail{}, err
	}
	if err := parseItiliumMutationResponse(payload); err != nil {
		return models.TicketDetail{}, err
	}

	requestedID := strings.TrimSpace(request.ResponsibleID)
	detail, err := c.pollTicketResponsibleAfterChange(ctx, request.UserID, number, requestedID)
	if err != nil {
		return models.TicketDetail{}, fmt.Errorf("poll ticket after change_responsible_sc: %w", err)
	}
	if requestedID != "" && strings.TrimSpace(detail.ResponsibleEmployeeID) != requestedID {
		c.logger.Warn(
			"find_sc returned stale responsible after change_responsible_sc",
			"number", number,
			"requested_responsible_id", requestedID,
			"requested_team_id", request.TeamID,
			"find_sc_responsible_id", detail.ResponsibleEmployeeID,
			"find_sc_responsible_title", detail.ResponsibleEmployee,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
		return models.TicketDetail{}, &ResponsibleChangeNotConfirmedError{
			Number:                 number,
			RequestedResponsibleID: requestedID,
			ActualResponsibleID:    detail.ResponsibleEmployeeID,
			ActualResponsibleTitle: detail.ResponsibleEmployee,
		}
	}
	return detail, nil
}

func responsibleOptionMatch(options []models.ResponsibleOption, responsibleID, teamID string) (models.ResponsibleOption, bool) {
	trimmedID := strings.TrimSpace(responsibleID)
	if trimmedID == "" {
		return models.ResponsibleOption{}, false
	}
	trimmedTeam := strings.TrimSpace(teamID)

	var fallback models.ResponsibleOption
	hasFallback := false
	for _, option := range options {
		if strings.TrimSpace(option.ExternalID) != trimmedID {
			continue
		}
		if trimmedTeam != "" && strings.TrimSpace(option.TeamExternalID) == trimmedTeam {
			return option, true
		}
		if !hasFallback {
			fallback = option
			hasFallback = true
		}
	}
	return fallback, hasFallback
}

// ConfirmTicket sends user rating via legacy confirm_sc (POST, query: id, telegram, incident, mark, optional comment_text).
func (c *Client) ConfirmTicket(ctx context.Context, number string, request models.ConfirmTicketRequest) (models.TicketDetail, error) {
	uid := strings.TrimSpace(request.UserID)
	q := url.Values{}
	q.Set("id", uid)
	q.Set("telegram", uid)
	q.Set("incident", strings.TrimSpace(number))
	q.Set("mark", strconv.Itoa(request.Mark))
	if comment := strings.TrimSpace(request.Comment); comment != "" {
		q.Set("comment_text", comment)
	}

	if _, err := c.doPostEmptyQuery(ctx, "/confirm_sc", q); err != nil {
		return models.TicketDetail{}, err
	}

	detail, err := c.GetTicket(ctx, request.UserID, number)
	if err != nil {
		return models.TicketDetail{Number: strings.TrimSpace(number)}, nil
	}
	return detail, nil
}

// SearchTicket searches a ticket by number.
func (c *Client) SearchTicket(ctx context.Context, request models.SearchTicketRequest) (models.TicketDetail, error) {
	// По контракту legacy поиск карточки выполняется через find_sc.
	return c.GetTicket(ctx, request.UserID, request.Number)
}

func parseFindSCResponse(payload map[string]any, fallbackNumber string) models.TicketDetail {
	if payload == nil {
		return models.TicketDetail{Number: fallbackNumber}
	}

	source := payload
	if nested, ok := payload["data"].(map[string]any); ok {
		source = nested
	}

	detail := models.TicketDetail{
		Number:               pickStringFromMap(source, "number", "Number", "sc_number", "sc", "ServiceCallNumber", "Номер"),
		Title:                pickStringFromMap(source, "title", "Title", "shortDescription", "ShortDescription", "Тема"),
		Description:          pickStringFromMap(source, "description", "Description", "text", "Текст"),
		CreationDate:         pickStringFromMap(source, "creationDate", "CreationDate", "dateCreate", "ДатаСоздания"),
		State:                pickStringFromMap(source, "state", "State", "status", "Status", "Состояние"),
		Deadline:             pickStringFromMap(source, "deadline", "Deadline", "deadlineDate", "executionDate", "ДатаИсполнения"),
		ResponsibleEmployee:   pickStringFromMap(source, "responsibleEmployee", "responsibleEmployeeTitle", "ResponsibleEmployeeTitle"),
		ResponsibleEmployeeID: pickStringFromMap(source, "responsibleEmployeeId", "ResponsibleEmployeeId", "responsibleId", "ResponsibleId"),
		ResponsibleTeam:       pickStringFromMap(source, "responsibleTeam", "responsibleTeamTitle", "ResponsibleTeam", "client", "OU", "Подразделение"),
		CanChangeStatus:      boolFromAny(firstAny(source, "canChangeStatus", "change_status")),
		CanChangeResponsible: boolFromAny(firstAny(source, "canChangeResponsible", "change_responsible")),
		CanConfirmRating:     boolFromAny(firstAny(source, "canConfirmRating", "needRating", "confirm_rating", "need_confirm")),
		AvailableStates:      firstStringSlice(source, "availableStates", "new_state", "newState"),
	}

	if detail.Number == "" {
		detail.Number = fallbackNumber
	}
	if detail.Title == "" {
		detail.Title = "Заявка " + detail.Number
	}

	return detail
}

func firstAny(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			return value
		}
	}
	return nil
}

func firstStringSlice(m map[string]any, keys ...string) []string {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			if items := stringSliceFromAny(value); len(items) > 0 {
				return items
			}
		}
	}
	return nil
}

// ListResponsibleOptions returns available assignees for the ticket.
// На контуре 1С параметры нужно передавать в query URL; GET часто даёт 405, multipart с telegram — 500, поэтому сначала POST без тела.
func (c *Client) ListResponsibleOptions(ctx context.Context, userID string, number string) ([]models.ResponsibleOption, error) {
	uid := strings.TrimSpace(userID)
	num := strings.TrimSpace(number)

	qID := url.Values{}
	qID.Set("id", uid)
	qID.Set("sc_number", num)

	payloadRaw, err := c.doPostEmptyQuery(ctx, "/responsibles_sc", qID)
	if err != nil {
		qLegacy := url.Values{}
		qLegacy.Set("telegram", uid)
		qLegacy.Set("sc_number", num)
		payloadRaw, err = c.doPostEmptyQuery(ctx, "/responsibles_sc", qLegacy)
	}
	if err != nil {
		// Последний шанс: multipart только id + sc_number (без telegram — иначе падает РаботаСМакс на части публикаций).
		form := url.Values{}
		form.Set("id", uid)
		form.Set("sc_number", num)
		payloadRaw, err = c.doMultipartFormPostBytes(ctx, "/responsibles_sc", form)
	}
	if err != nil {
		return nil, err
	}

	return parseResponsiblesScResponse(payloadRaw)
}

// doJSONRequestPayload performs the same HTTP exchange as doJSON and returns the response body
// (after 2xx/3xx; 4xx/5xx return HTTPStatusError and no payload is unmarshaled by callers that need a body on success).
func (c *Client) doJSONRequestPayload(ctx context.Context, method string, path string, query map[string]string, requestBody any) ([]byte, error) {
	if c.baseURL == "" {
		return nil, errors.New("itilium base url is required")
	}

	// baseURL уже без завершающего слэша; path начинается с /…
	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse itilium url: %w", err)
	}

	values := endpoint.Query()
	for key, value := range query {
		values.Set(key, value)
	}
	endpoint.RawQuery = values.Encode()

	var body io.Reader
	var requestRaw []byte
	if requestBody != nil {
		requestRaw, err = json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewBuffer(requestRaw)
	}

	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	if c.login != "" {
		request.SetBasicAuth(c.login, c.password)
	}

	start := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error(
			"itilium request failed",
			"method", method,
			"url", endpoint.String(),
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	c.logger.Info(
		"itilium request completed",
		"method", method,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)
	c.logger.Debug(
		"itilium request details",
		"method", method,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"request_body", string(requestRaw),
		"response_body", string(payload),
		"request_id", middleware.RequestIDFromContext(ctx),
		"user_id", middleware.UserIDFromContext(ctx),
	)

	// 4xx/5xx отдаём сервису как HTTPStatusError — там решается UI (регистрация, ожидание и т.д.).
	if response.StatusCode >= 400 {
		return nil, HTTPStatusError{StatusCode: response.StatusCode}
	}

	return payload, nil
}

// unmarshalItiliumObjectOrStringError decodes 1С JSON: либо объект карточки, либо JSON-строка с сообщением
// (например "Заявка с таким номером не найдена" при HTTP 200).
func unmarshalItiliumObjectOrStringError(payload []byte) (map[string]any, error) {
	if len(bytes.TrimSpace(payload)) == 0 {
		return nil, nil
	}
	var v any
	if err := json.Unmarshal(payload, &v); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if v == nil {
		return nil, nil
	}
	if s, ok := v.(string); ok {
		return nil, errors.New(s)
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("decode response: expected object or string, got %T", v)
	}
	return m, nil
}

// doJSON performs a JSON request and logs request and response metadata for troubleshooting.
func (c *Client) doJSON(ctx context.Context, method string, path string, query map[string]string, requestBody any, responseBody any) error {
	payload, err := c.doJSONRequestPayload(ctx, method, path, query, requestBody)
	if err != nil {
		return err
	}

	if responseBody == nil || len(payload) == 0 {
		return nil
	}

	if err := json.Unmarshal(payload, responseBody); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// doFormPostBytes выполняет POST с телом application/x-www-form-urlencoded и возвращает сырое тело ответа.
func (c *Client) doFormPostBytes(ctx context.Context, path string, form url.Values) ([]byte, error) {
	if c.baseURL == "" {
		return nil, errors.New("itilium base url is required")
	}

	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse itilium url: %w", err)
	}

	encodedForm := form.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(encodedForm))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.login != "" {
		request.SetBasicAuth(c.login, c.password)
	}

	start := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error("itilium form request failed", "method", http.MethodPost, "url", endpoint.String(), "duration_ms", time.Since(start).Milliseconds(), "error", err)
		return nil, fmt.Errorf("perform form request: %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// urlencoded_body — копия тела запроса для логов; в 1С уходят обычные имена полей формы (id, shortDescription, description).
	respText := strings.TrimPrefix(string(payload), "\ufeff")
	c.logger.Info(
		"itilium form request completed",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"urlencoded_body", truncateLogString(encodedForm, 12000),
		"response_body", truncateLogString(respText, 8000),
	)
	c.logger.Debug(
		"itilium form request details",
		"method", http.MethodPost,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"form_body", encodedForm,
		"response_body", respText,
	)

	if response.StatusCode >= 400 {
		return nil, HTTPStatusError{StatusCode: response.StatusCode}
	}

	return payload, nil
}

// doFormPostWithGet405FallbackBytes отправляет POST как в legacy-контрактах; если 1С публикует только GET
// (типичная ситуация: 405 Method Not Allowed), повторяет с теми же полями в query-string.
func (c *Client) doFormPostWithGet405FallbackBytes(ctx context.Context, path string, form url.Values) ([]byte, error) {
	payload, err := c.doFormPostBytes(ctx, path, form)
	if err == nil {
		return payload, nil
	}
	var hs HTTPStatusError
	if errors.As(err, &hs) && hs.StatusCode == http.StatusMethodNotAllowed {
		c.logger.Info(
			"itilium POST returned 405 on form endpoint, retrying GET with same fields as query",
			"path", path,
			"request_id", middleware.RequestIDFromContext(ctx),
			"user_id", middleware.UserIDFromContext(ctx),
		)
		return c.doFormGetBytes(ctx, path, form)
	}
	return nil, err
}

// doFormGetBytes — GET с параметрами формы в query (идентичные имена полей, что и в urlencoded теле POST).
func (c *Client) doFormGetBytes(ctx context.Context, path string, form url.Values) ([]byte, error) {
	if c.baseURL == "" {
		return nil, errors.New("itilium base url is required")
	}

	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse itilium url: %w", err)
	}
	endpoint.RawQuery = form.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if c.login != "" {
		request.SetBasicAuth(c.login, c.password)
	}

	start := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error("itilium form GET request failed", "method", http.MethodGet, "url", endpoint.String(), "duration_ms", time.Since(start).Milliseconds(), "error", err)
		return nil, fmt.Errorf("perform form GET request: %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	respText := strings.TrimPrefix(string(payload), "\ufeff")
	c.logger.Info(
		"itilium form GET request completed",
		"method", http.MethodGet,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"query_string", truncateLogString(form.Encode(), 12000),
		"response_body", truncateLogString(respText, 8000),
	)
	c.logger.Debug(
		"itilium form GET request details",
		"method", http.MethodGet,
		"url", endpoint.String(),
		"status_code", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"query", form.Encode(),
		"response_body", respText,
	)

	if response.StatusCode >= 400 {
		return nil, HTTPStatusError{StatusCode: response.StatusCode}
	}

	return payload, nil
}

// doForm performs a x-www-form-urlencoded request for legacy ITILIUM endpoints.
func (c *Client) doForm(ctx context.Context, method string, path string, form url.Values, responseBody any) error {
	if c.baseURL == "" {
		return errors.New("itilium base url is required")
	}

	// Тело — application/x-www-form-urlencoded, Basic Auth как в doJSON.
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

	// Та же семантика, что в doJSON: ошибочный статус — для ветвления в ProfileService.
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
	// В демо-режиме в таймлайн выводим факт наличия вложений по именам (как пришли с multipart).
	msg := "Заявка создана и отправлена в ITILIUM."
	if len(request.Attachments) > 0 {
		msg = fmt.Sprintf("Заявка создана. Вложения: %s.", strings.Join(request.Attachments, ", "))
	}

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
			{Author: "Система", Message: msg, CreatedAt: time.Now().Format(time.RFC3339)},
		},
	}, nil
}

// AddComment appends a synthetic comment to the static ticket.
func (c *DemoClient) AddComment(_ context.Context, number string, request models.AddCommentRequest) (models.TicketDetail, error) {
	ticket := demoTicket(number)
	msg := strings.TrimSpace(request.Message)
	if msg == "" && len(request.FileAttachments) > 0 {
		names := make([]string, 0, len(request.FileAttachments))
		for _, fa := range request.FileAttachments {
			if strings.TrimSpace(fa.Filename) != "" {
				names = append(names, fa.Filename)
			}
		}
		msg = "Прикреплён файл."
		if len(names) > 0 {
			msg = fmt.Sprintf("Вложения: %s.", strings.Join(names, ", "))
		}
	}
	ticket.Timeline = append(ticket.Timeline, models.CommentEntry{
		Author:    "Пользователь",
		Message:   msg,
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

// ConfirmTicket records a synthetic rating on the demo ticket.
func (c *DemoClient) ConfirmTicket(_ context.Context, number string, request models.ConfirmTicketRequest) (models.TicketDetail, error) {
	ticket := demoTicket(number)
	ticket.Timeline = append(ticket.Timeline, models.CommentEntry{
		Author:    "Пользователь",
		Message:   fmt.Sprintf("Оценка отправлена: %d. %s", request.Mark, strings.TrimSpace(request.Comment)),
		CreatedAt: time.Now().Format(time.RFC3339),
	})
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

// ListMarketingServices returns deterministic marketing types with dynamic form schemas.
func (c *DemoClient) ListMarketingServices(_ context.Context, _ string) ([]models.MarketingServiceType, error) {
	return []models.MarketingServiceType{
		{Code: "design", Name: "Дизайн", FormNumber: "1", FormSchema: defaultMarketingFormSchema("1", "Дизайн")},
		{Code: "event", Name: "Мероприятие", FormNumber: "2", FormSchema: defaultMarketingFormSchema("2", "Мероприятие")},
		{Code: "ads", Name: "Реклама", FormNumber: "3", FormSchema: defaultMarketingFormSchema("3", "Реклама")},
		{Code: "smm", Name: "SMM", FormNumber: "4", FormSchema: defaultMarketingFormSchema("4", "SMM")},
		{Code: "promo", Name: "Акция", FormNumber: "5", FormSchema: defaultMarketingFormSchema("5", "Акция")},
		{Code: "other", Name: "Иное", FormNumber: "6", FormSchema: defaultMarketingFormSchema("6", "Иное")},
	}, nil
}

// ListMarketingSubdivisions returns deterministic marketing subdivisions for demo mode.
func (c *DemoClient) ListMarketingSubdivisions(_ context.Context, _ string) ([]models.MarketingSubdivision, error) {
	return []models.MarketingSubdivision{
		{Name: "БП – Барская Пивница"},
		{Name: "ИВ – Иван Васильевич"},
		{Name: "ВС – Время Счастья"},
		{Name: "СБ – супермаркет Барс"},
	}, nil
}

// CreateMarketingRequest returns a synthetic ticket after dynamic marketing form submit.
func (c *DemoClient) CreateMarketingRequest(_ context.Context, request models.CreateMarketingRequest) (models.TicketDetail, error) {
	return models.TicketDetail{
		Number:               "SC-MKT-001",
		Title:                "Маркетинговая заявка: " + request.ServiceCode,
		Description:          buildMarketingDescription(request),
		State:                "Зарегистрирована",
		Deadline:             request.ExecutionDate,
		ResponsibleTeam:      request.Subdivision,
		CanChangeResponsible: true,
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
		State:                "08_Закрыто",
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

// truncateLogString ограничивает длину строк в логах (UTF-8 может обрезаться по байтам — для логов допустимо).
func truncateLogString(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes] + fmt.Sprintf(" … (truncated, %d bytes total)", len(s))
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
