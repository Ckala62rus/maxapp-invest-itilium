// Package handlers contains HTTP handlers and route wiring.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/auth"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/middleware"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/services"
)

// Handler groups the service dependencies used by HTTP endpoints.
type Handler struct {
	logger         *slog.Logger
	authService    *services.AuthService
	authManager    *auth.Manager
	debugIdentity  bool
	profileService *services.ProfileService
	ticketService  *services.TicketService
}

// New creates a handler bundle.
func New(
	logger *slog.Logger,
	authService *services.AuthService,
	authManager *auth.Manager,
	debugIdentity bool,
	profileService *services.ProfileService,
	ticketService *services.TicketService,
) *Handler {
	return &Handler{
		logger:         logger,
		authService:    authService,
		authManager:    authManager,
		debugIdentity:  debugIdentity,
		profileService: profileService,
		ticketService:  ticketService,
	}
}

// Health returns a basic liveness response.
func (h *Handler) Health(writer http.ResponseWriter, request *http.Request) {
	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "ok",
	})
}

// Ready returns a basic readiness response.
func (h *Handler) Ready(writer http.ResponseWriter, request *http.Request) {
	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "ready",
	})
}

// GetProfile returns the current user profile.
func (h *Handler) GetProfile(writer http.ResponseWriter, request *http.Request) {
	// userId уже положен middleware Identity (из Bearer или debug).
	profile, err := h.profileService.GetProfile(request.Context(), middleware.UserIDFromContext(request.Context()))
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    profile,
	})
}

// RegisterUser stores the registration request from the UI.
func (h *Handler) RegisterUser(writer http.ResponseWriter, request *http.Request) {
	var payload models.RegistrationRequest
	if err := decodeJSON(request, &payload); err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	// Если фронт не прислал userId в теле — берём из контекста (тот же MAX id, что и в токене).
	if payload.UserID == "" {
		payload.UserID = middleware.UserIDFromContext(request.Context())
	}

	profile, err := h.profileService.Register(request.Context(), payload)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "registration request sent for review",
		Data:    profile,
	})
}

// ValidateMaxAuth exchanges signed MAX initData for a backend access token.
func (h *Handler) ValidateMaxAuth(writer http.ResponseWriter, request *http.Request) {
	var payload models.MaxAuthValidateRequest
	if err := decodeJSON(request, &payload); err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	// Логируем только длину строки — сам initData в лог не пишем (подписанные данные).
	h.logger.Info(
		"max auth request received",
		"request_id", middleware.RequestIDFromContext(request.Context()),
		"path", request.URL.Path,
		"init_data_length", len(payload.InitData),
	)

	response, err := h.authService.ValidateMaxInitData(request.Context(), payload.InitData)
	if err != nil {
		// Неверная подпись, просрочка auth_date и т.д. — с точки зрения HTTP это «не авторизован».
		h.writeError(writer, request, http.StatusUnauthorized, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "max auth validated",
		Data:    response,
	})
}

// FindEmployeeByIdentifier calls the legacy ITILIUM employee lookup endpoint through the service layer.
func (h *Handler) FindEmployeeByIdentifier(writer http.ResponseWriter, request *http.Request) {
	var payload models.EmployeeLookupRequest
	if err := decodeJSON(request, &payload); err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	// Удобно для отладки: пустой identifier → текущий пользователь из токена.
	if payload.Identifier == "" {
		payload.Identifier = middleware.UserIDFromContext(request.Context())
	}

	employee, err := h.profileService.FindEmployeeByIdentifier(request.Context(), payload)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "employee payload loaded",
		Data:    employee,
	})
}

// ListMyTickets returns the current user's own tickets.
func (h *Handler) ListMyTickets(writer http.ResponseWriter, request *http.Request) {
	// Список «мои заявки» в ITILIUM по userId из токена.
	tickets, err := h.ticketService.ListMyTickets(request.Context(), middleware.UserIDFromContext(request.Context()))
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{Success: true, Data: tickets})
}

// ListResponsibleTickets returns tickets assigned to the current user.
func (h *Handler) ListResponsibleTickets(writer http.ResponseWriter, request *http.Request) {
	// Заявки, где пользователь в роли ответственного.
	tickets, err := h.ticketService.ListResponsibleTickets(request.Context(), middleware.UserIDFromContext(request.Context()))
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{Success: true, Data: tickets})
}

// SearchTicket searches a ticket by number.
func (h *Handler) SearchTicket(writer http.ResponseWriter, request *http.Request) {
	var payload models.SearchTicketRequest
	if err := decodeJSON(request, &payload); err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	// Инициатор поиска — текущий пользователь (для фильтрации на стороне ITILIUM).
	if payload.UserID == "" {
		payload.UserID = middleware.UserIDFromContext(request.Context())
	}

	ticket, err := h.ticketService.SearchTicket(request.Context(), payload)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{Success: true, Data: ticket})
}

// Лимиты тела multipart при создании заявки: защита от OOM и злоупотреблений по числу/размеру файлов.
const (
	maxCreateTicketMultipartMemory = 32 << 20 // весь multipart в памяти до ~32 MiB
	maxAttachmentSize              = 15 << 20 // один файл не больше ~15 MiB
	maxAttachmentCount             = 20        // не больше 20 вложений за один запрос
)

// CreateTicket creates a new ticket through the service layer.
func (h *Handler) CreateTicket(writer http.ResponseWriter, request *http.Request) {
	var payload models.CreateTicketRequest

	ct := request.Header.Get("Content-Type")
	// Фронт с вложениями шлёт multipart: поле payload (JSON) + повторяющиеся части attachments (файлы).
	// Без файлов остаётся прежний JSON — совместимость со старыми клиентами.
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := request.ParseMultipartForm(maxCreateTicketMultipartMemory); err != nil {
			h.writeError(writer, request, http.StatusBadRequest, err)
			return
		}

		payloadStr := request.FormValue("payload")
		if strings.TrimSpace(payloadStr) == "" {
			h.writeError(writer, request, http.StatusBadRequest, errors.New("payload field is required for multipart create"))
			return
		}

		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			h.writeError(writer, request, http.StatusBadRequest, err)
			return
		}

		if request.MultipartForm == nil {
			h.writeError(writer, request, http.StatusBadRequest, errors.New("multipart form is empty"))
			return
		}

		files := request.MultipartForm.File["attachments"]
		if len(files) > maxAttachmentCount {
			h.writeError(writer, request, http.StatusBadRequest, fmt.Errorf("too many attachments (max %d)", maxAttachmentCount))
			return
		}

		// Имена и сырые байты заполняем заново из фактически загруженных частей формы.
		payload.Attachments = nil
		payload.FileAttachments = nil

		for _, fh := range files {
			if fh.Size > maxAttachmentSize {
				h.writeError(writer, request, http.StatusBadRequest, fmt.Errorf("attachment %q exceeds size limit", fh.Filename))
				return
			}

			f, err := fh.Open()
			if err != nil {
				h.writeError(writer, request, http.StatusBadRequest, err)
				return
			}

			// LimitReader(+1): если прочитали больше лимита — отклоняем даже при заниженном fh.Size.
			data, err := io.ReadAll(io.LimitReader(f, maxAttachmentSize+1))
			_ = f.Close()
			if err != nil {
				h.writeError(writer, request, http.StatusBadRequest, err)
				return
			}

			if int64(len(data)) > maxAttachmentSize {
				h.writeError(writer, request, http.StatusBadRequest, fmt.Errorf("attachment %q exceeds size limit", fh.Filename))
				return
			}

			ctype := fh.Header.Get("Content-Type")
			if ctype == "" {
				ctype = http.DetectContentType(data)
			}

			payload.FileAttachments = append(payload.FileAttachments, models.FileAttachment{
				Filename:    fh.Filename,
				ContentType: ctype,
				Data:        data,
			})
			payload.Attachments = append(payload.Attachments, fh.Filename)
		}
	} else {
		// Обычное создание без файлов: одно тело application/json.
		if err := decodeJSON(request, &payload); err != nil {
			h.writeError(writer, request, http.StatusBadRequest, err)
			return
		}
	}

	// Автор заявки — текущий пользователь, если не передан явно.
	if payload.UserID == "" {
		payload.UserID = middleware.UserIDFromContext(request.Context())
	}

	ticket, err := h.ticketService.CreateTicket(request.Context(), payload)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "ticket created",
		Data:    ticket,
	})
}

// GetTicket returns the ticket detail page payload.
func (h *Handler) GetTicket(writer http.ResponseWriter, request *http.Request) {
	number := request.PathValue("number")
	// Карточка с кэшем в TicketService (Redis), ключ включает userId и номер заявки.
	ticket, err := h.ticketService.GetTicket(request.Context(), middleware.UserIDFromContext(request.Context()), number)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{Success: true, Data: ticket})
}

// AddComment appends a comment to the selected ticket.
func (h *Handler) AddComment(writer http.ResponseWriter, request *http.Request) {
	var payload models.AddCommentRequest

	ct := request.Header.Get("Content-Type")
	// С вложениями — multipart: поле payload (JSON) + повторяющиеся части attachments (как при создании заявки).
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := request.ParseMultipartForm(maxCreateTicketMultipartMemory); err != nil {
			h.writeError(writer, request, http.StatusBadRequest, err)
			return
		}

		payloadStr := request.FormValue("payload")
		if strings.TrimSpace(payloadStr) == "" {
			h.writeError(writer, request, http.StatusBadRequest, errors.New("payload field is required for multipart comment"))
			return
		}

		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			h.writeError(writer, request, http.StatusBadRequest, err)
			return
		}

		if request.MultipartForm == nil {
			h.writeError(writer, request, http.StatusBadRequest, errors.New("multipart form is empty"))
			return
		}

		files := request.MultipartForm.File["attachments"]
		if len(files) > maxAttachmentCount {
			h.writeError(writer, request, http.StatusBadRequest, fmt.Errorf("too many attachments (max %d)", maxAttachmentCount))
			return
		}

		payload.Attachments = nil
		payload.FileAttachments = nil

		for _, fh := range files {
			if fh.Size > maxAttachmentSize {
				h.writeError(writer, request, http.StatusBadRequest, fmt.Errorf("attachment %q exceeds size limit", fh.Filename))
				return
			}

			f, err := fh.Open()
			if err != nil {
				h.writeError(writer, request, http.StatusBadRequest, err)
				return
			}

			data, err := io.ReadAll(io.LimitReader(f, maxAttachmentSize+1))
			_ = f.Close()
			if err != nil {
				h.writeError(writer, request, http.StatusBadRequest, err)
				return
			}

			if int64(len(data)) > maxAttachmentSize {
				h.writeError(writer, request, http.StatusBadRequest, fmt.Errorf("attachment %q exceeds size limit", fh.Filename))
				return
			}

			ctype := fh.Header.Get("Content-Type")
			if ctype == "" {
				ctype = http.DetectContentType(data)
			}

			payload.FileAttachments = append(payload.FileAttachments, models.FileAttachment{
				Filename:    fh.Filename,
				ContentType: ctype,
				Data:        data,
			})
			payload.Attachments = append(payload.Attachments, fh.Filename)
		}
	} else {
		if err := decodeJSON(request, &payload); err != nil {
			h.writeError(writer, request, http.StatusBadRequest, err)
			return
		}
	}

	// Кто пишет комментарий — всегда из сессии (не доверяем телу запроса).
	payload.UserID = middleware.UserIDFromContext(request.Context())
	ticket, err := h.ticketService.AddComment(request.Context(), request.PathValue("number"), payload)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{Success: true, Message: "comment added", Data: ticket})
}

// ChangeStatus updates the ticket workflow state.
func (h *Handler) ChangeStatus(writer http.ResponseWriter, request *http.Request) {
	var payload models.ChangeStatusRequest
	if err := decodeJSON(request, &payload); err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	payload.UserID = middleware.UserIDFromContext(request.Context())
	ticket, err := h.ticketService.ChangeStatus(request.Context(), request.PathValue("number"), payload)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{Success: true, Message: "status changed", Data: ticket})
}

// ChangeResponsible updates the current responsible person.
func (h *Handler) ChangeResponsible(writer http.ResponseWriter, request *http.Request) {
	var payload models.ChangeResponsibleRequest
	if err := decodeJSON(request, &payload); err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	payload.UserID = middleware.UserIDFromContext(request.Context())
	ticket, err := h.ticketService.ChangeResponsible(request.Context(), request.PathValue("number"), payload)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{Success: true, Message: "responsible changed", Data: ticket})
}

// ConfirmTicket sends resolution rating to ITILIUM (confirm_sc).
func (h *Handler) ConfirmTicket(writer http.ResponseWriter, request *http.Request) {
	var payload models.ConfirmTicketRequest
	if err := decodeJSON(request, &payload); err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	payload.UserID = middleware.UserIDFromContext(request.Context())
	ticket, err := h.ticketService.ConfirmTicket(request.Context(), request.PathValue("number"), payload)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{Success: true, Message: "rating submitted", Data: ticket})
}

// ListResponsibleOptions returns available assignees for the selected ticket.
func (h *Handler) ListResponsibleOptions(writer http.ResponseWriter, request *http.Request) {
	// Справочник для смены ответственного по {number} в пути.
	options, err := h.ticketService.ListResponsibleOptions(
		request.Context(),
		middleware.UserIDFromContext(request.Context()),
		request.PathValue("number"),
	)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err)
		return
	}

	h.writeJSON(writer, request, http.StatusOK, models.APIResponse{Success: true, Data: options})
}

// writeJSON returns a successful JSON payload with the current request id.
func (h *Handler) writeJSON(writer http.ResponseWriter, request *http.Request, status int, payload models.APIResponse) {
	// requestId в теле дублирует заголовок — удобно для фронта и поддержки.
	payload.RequestID = middleware.RequestIDFromContext(request.Context())
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}

// writeError returns a structured error response.
func (h *Handler) writeError(writer http.ResponseWriter, request *http.Request, status int, err error) {
	// Текст ошибки уходит клиенту в message — не добавляйте сюда секреты.
	h.logger.Warn(
		"http handler returned error",
		"request_id", middleware.RequestIDFromContext(request.Context()),
		"user_id", middleware.UserIDFromContext(request.Context()),
		"path", request.URL.Path,
		"error", err,
	)

	h.writeJSON(writer, request, status, models.APIResponse{
		Success: false,
		Message: err.Error(),
	})
}

// decodeJSON decodes a JSON request body into the provided target.
func decodeJSON(request *http.Request, target any) error {
	if request.Body == nil {
		return errors.New("request body is required")
	}

	return json.NewDecoder(request.Body).Decode(target)
}
