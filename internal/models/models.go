// Package models stores transport and domain structures shared across layers.
package models

// UserProfile describes the current MAX user linked to ITILIUM.
type UserProfile struct {
	// UserID stores the external MAX user identifier.
	UserID string `json:"userId"`
	// Username stores the user handle shown in the UI.
	Username string `json:"username"`
	// FullName stores the display name used in cards and forms.
	FullName string `json:"fullName"`
	// FirstName stores the surname/last name part returned by ITILIUM when available.
	FirstName string `json:"firstName,omitempty"`
	// LastName stores the given name part returned by ITILIUM when available.
	LastName string `json:"lastName,omitempty"`
	// MiddleName stores the patronymic or middle name returned by ITILIUM.
	MiddleName string `json:"middleName,omitempty"`
	// Department stores the current department or branch.
	Department string `json:"department"`
	// Organization stores the legal entity or organization name.
	Organization string `json:"organization,omitempty"`
	// Position stores the user's job title in ITILIUM.
	Position string `json:"position,omitempty"`
	// ServiceCalls stores ticket numbers already linked to the employee.
	ServiceCalls []string `json:"servicecalls,omitempty"`
	// CanCreateMarketingRequests tells the UI whether marketing request flows should be available.
	CanCreateMarketingRequests bool `json:"canCreateMarketingRequests,omitempty"`
	// CanCreateDaxRequests tells the UI whether DAX request flows should be available.
	CanCreateDaxRequests bool `json:"canCreateDaxRequests"`
	// EmployeeFound shows whether ITILIUM already knows this user.
	EmployeeFound bool `json:"employeeFound"`
	// RegistrationRequired tells the UI to open the registration form.
	RegistrationRequired bool `json:"registrationRequired"`
	// RegistrationPending shows that a manual registration request already exists and is still under review.
	RegistrationPending bool `json:"registrationPending"`
	// StatusMessage stores a user-facing status hint for onboarding and review states.
	StatusMessage string `json:"statusMessage,omitempty"`
}

// MaxAuthValidateRequest stores the raw MAX initData sent from the mini app.
type MaxAuthValidateRequest struct {
	// InitData stores the signed MAX WebAppData string from window.WebApp.initData.
	InitData string `json:"initData"`
}

// MaxIdentity stores the trusted MAX user identity returned after validation.
type MaxIdentity struct {
	// UserID stores the trusted MAX user id.
	UserID string `json:"userId"`
	// Username stores the MAX username or nickname when available.
	Username string `json:"username,omitempty"`
	// FullName stores the display name assembled from MAX identity fields.
	FullName string `json:"fullName,omitempty"`
	// FirstName stores the original MAX first name field.
	FirstName string `json:"firstName,omitempty"`
	// LastName stores the original MAX last name field.
	LastName string `json:"lastName,omitempty"`
}

// MaxAuthValidateResponse stores the backend token and trusted MAX identity.
type MaxAuthValidateResponse struct {
	// AccessToken stores the backend bearer token used for the rest of the API.
	AccessToken string `json:"accessToken"`
	// ExpiresAt stores the unix timestamp when the backend token expires.
	ExpiresAt int64 `json:"expiresAt"`
	// Identity stores the trusted MAX identity derived from validated initData.
	Identity MaxIdentity `json:"identity"`
}

// RegistrationRequest describes the form used when the user is missing in ITILIUM.
type RegistrationRequest struct {
	// UserID stores the external MAX user identifier.
	UserID string `json:"userId"`
	// EmployeeNumber stores the user's internal employee number.
	EmployeeNumber string `json:"employeeNumber"`
	// FullName stores the user's full legal name.
	FullName string `json:"fullName"`
	// Organization stores the legal entity or organization name for registration.
	Organization string `json:"organization"`
	// Department stores the user's store, branch or department.
	Department string `json:"department"`
	// Position stores the user's job title.
	Position string `json:"position"`
	// Phone stores the callback phone number.
	Phone string `json:"phone"`
	// Comment stores free-form extra context.
	Comment string `json:"comment"`
}

// EmployeeLookupRequest describes a legacy-style user lookup in ITILIUM.
type EmployeeLookupRequest struct {
	// Identifier stores the value sent to ITILIUM, for example MAX user id.
	Identifier string `json:"identifier"`
	// AttributeCode stores the ITILIUM field name used for lookup, for example id or telegram.
	AttributeCode string `json:"attributeCode"`
}

// EmployeeLookupResult stores a normalized response from the legacy find_employee endpoint.
type EmployeeLookupResult struct {
	// Identifier stores the lookup value that was sent to ITILIUM.
	Identifier string `json:"identifier"`
	// AttributeCode stores the attribute name used for lookup.
	AttributeCode string `json:"attributeCode"`
	// UUID stores the employee UUID returned by ITILIUM.
	UUID string `json:"uuid"`
	// ServiceCalls stores the list of the user's ticket numbers from ITILIUM.
	ServiceCalls []string `json:"servicecalls"`
	// CanCreateMarketingRequests stores the marketing access flag returned by ITILIUM.
	CanCreateMarketingRequests bool `json:"canCreateMarketingRequests"`
	// CanCreateDaxRequests stores the DAX access flag returned by ITILIUM.
	CanCreateDaxRequests bool `json:"canCreateDaxRequests"`
	// Raw stores the full ITILIUM payload so unknown flags are not lost during exploration.
	Raw map[string]any `json:"raw"`
}

// TicketSummary describes a short card representation of a ticket.
type TicketSummary struct {
	// Number stores the ITILIUM ticket number.
	Number string `json:"number"`
	// Title stores the short description of the ticket.
	Title string `json:"title"`
	// State stores the current workflow state.
	State string `json:"state"`
	// Deadline stores the target completion date.
	Deadline string `json:"deadline"`
	// ResponsibleTeam stores the current responsible team.
	ResponsibleTeam string `json:"responsibleTeam"`
}

// CommentEntry describes a timeline event in the ticket card.
type CommentEntry struct {
	// Author stores the actor who created the event.
	Author string `json:"author"`
	// Message stores the visible timeline text.
	Message string `json:"message"`
	// CreatedAt stores the event time.
	CreatedAt string `json:"createdAt"`
}

// TicketDetail describes the full ticket page.
type TicketDetail struct {
	// Number stores the ITILIUM ticket number.
	Number string `json:"number"`
	// Title stores the short ticket subject.
	Title string `json:"title"`
	// Description stores the detailed ticket text.
	Description string `json:"description"`
	// CreationDate stores the ticket creation date from ITILIUM.
	CreationDate string `json:"creationDate,omitempty"`
	// State stores the current workflow state.
	State string `json:"state"`
	// Deadline stores the current SLA or requested date.
	Deadline string `json:"deadline"`
	// ResponsibleEmployee stores the current responsible employee display name.
	ResponsibleEmployee string `json:"responsibleEmployee,omitempty"`
	// ResponsibleTeam stores the team currently assigned to the ticket.
	ResponsibleTeam string `json:"responsibleTeam"`
	// CanChangeStatus shows whether status transition is allowed.
	CanChangeStatus bool `json:"canChangeStatus"`
	// CanChangeResponsible shows whether responsible reassignment is allowed.
	CanChangeResponsible bool `json:"canChangeResponsible"`
	// CanConfirmRating tells the UI to offer the post-resolution rating flow (confirm_sc).
	CanConfirmRating bool `json:"canConfirmRating,omitempty"`
	// AvailableStates lists next state transition names.
	AvailableStates []string `json:"availableStates"`
	// Timeline stores comments and important system events.
	Timeline []CommentEntry `json:"timeline"`
}

// FileAttachment holds one uploaded file for create-ticket multipart flows (not sent as JSON to ITILIUM).
// Один файл из multipart при создании заявки; в JSON API не сериализуется, только для проксирования в ITILIUM.
type FileAttachment struct {
	// Filename stores the original file name from the client.
	Filename string
	// ContentType stores the MIME type; may be empty and filled server-side.
	ContentType string
	// Data stores the raw file bytes.
	Data []byte
}

// MarketingFormField describes one dynamic field from a marketing form schema.
type MarketingFormField struct {
	// Key stores the machine-readable field identifier.
	Key string `json:"key"`
	// Label stores the user-facing field label.
	Label string `json:"label"`
	// Type stores UI control type (text, textarea, select, number, date, links).
	Type string `json:"type"`
	// Required tells whether the field must be filled.
	Required bool `json:"required"`
	// Placeholder stores optional UI hint inside the control.
	Placeholder string `json:"placeholder,omitempty"`
	// Hint stores optional helper text below the control.
	Hint string `json:"hint,omitempty"`
	// Options stores selectable values for select-like fields.
	Options []string `json:"options,omitempty"`
}

// MarketingFormSchema stores a normalized dynamic schema identified by ITILIUM form number.
type MarketingFormSchema struct {
	// FormNumber stores the 1C form identifier used by the frontend renderer.
	FormNumber string `json:"formNumber"`
	// Title stores the human-readable schema title.
	Title string `json:"title,omitempty"`
	// Fields stores all dynamic controls required by the selected marketing type.
	Fields []MarketingFormField `json:"fields"`
}

// MarketingServiceType describes one selectable marketing request type from ITILIUM.
type MarketingServiceType struct {
	// Code stores the internal type code used by ITILIUM.
	Code string `json:"code"`
	// Name stores the display name shown in the UI.
	Name string `json:"name"`
	// FormNumber stores the dynamic form number associated with this type.
	FormNumber string `json:"formNumber"`
	// FormSchema stores the normalized dynamic field definition for step 4.
	FormSchema MarketingFormSchema `json:"formSchema"`
}

// MarketingSubdivision describes one selectable subdivision for marketing requests.
type MarketingSubdivision struct {
	// Code stores the subdivision identifier when available.
	Code string `json:"code,omitempty"`
	// Name stores the subdivision display label.
	Name string `json:"name"`
}

// CreateMarketingRequest stores the payload for a 4-step marketing ticket flow.
type CreateMarketingRequest struct {
	// UserID stores the acting MAX user identifier.
	UserID string `json:"userId"`
	// ServiceCode stores selected marketing service/type code from ITILIUM.
	ServiceCode string `json:"serviceCode"`
	// FormNumber stores selected dynamic form number from ITILIUM.
	FormNumber string `json:"formNumber"`
	// Subdivision stores selected marketing subdivision.
	Subdivision string `json:"subdivision"`
	// ExecutionDate stores requested execution date; may be empty when date is omitted.
	ExecutionDate string `json:"executionDate"`
	// WithoutDate stores explicit "without date" flag from the UI.
	WithoutDate bool `json:"withoutDate"`
	// FormData stores dynamic key/value map filled on step 4.
	FormData map[string]string `json:"formData"`
	// Attachments stores uploaded file names or references.
	Attachments []string `json:"attachments"`
	// FileAttachments stores raw uploads when the client sends multipart/form-data.
	FileAttachments []FileAttachment `json:"-"`
}

// CreateTicketRequest stores the payload used to create a new ticket.
type CreateTicketRequest struct {
	// UserID stores the acting MAX user identifier.
	UserID string `json:"userId"`
	// RequestType stores the selected ticket type.
	RequestType string `json:"requestType"`
	// Title stores the short subject of the request.
	Title string `json:"title"`
	// Description stores the detailed request body.
	Description string `json:"description"`
	// Department stores the selected business department.
	Department string `json:"department"`
	// ExecutionDate stores the requested completion date.
	ExecutionDate string `json:"executionDate"`
	// Attachments stores uploaded file names or references.
	Attachments []string `json:"attachments"`
	// FileAttachments stores raw uploads when the client sends multipart/form-data.
	// Сырые вложения (multipart); в JSON наружу не отдаём — только для внутренней передачи в клиент ITILIUM.
	FileAttachments []FileAttachment `json:"-"`
}

// SearchTicketRequest stores the search input from the UI.
type SearchTicketRequest struct {
	// Number stores the ticket number entered by the user.
	Number string `json:"number"`
	// UserID stores the acting MAX user identifier.
	UserID string `json:"userId"`
}

// AddCommentRequest stores a new comment payload for a ticket.
type AddCommentRequest struct {
	// UserID stores the acting MAX user identifier.
	UserID string `json:"userId"`
	// Message stores the comment body.
	Message string `json:"message"`
	// Attachments stores uploaded file names or references.
	Attachments []string `json:"attachments"`
	// FileAttachments stores raw uploads when the client sends multipart/form-data (same pattern as create ticket).
	FileAttachments []FileAttachment `json:"-"`
}

// ChangeStatusRequest stores a workflow transition payload.
type ChangeStatusRequest struct {
	// UserID stores the acting MAX user identifier.
	UserID string `json:"userId"`
	// State stores the target workflow state.
	State string `json:"state"`
	// Comment stores optional transition commentary.
	Comment string `json:"comment"`
	// Date stores an optional deferral or target date.
	Date string `json:"date"`
}

// ChangeResponsibleRequest stores the selected new responsible person.
type ChangeResponsibleRequest struct {
	// UserID stores the acting MAX user identifier.
	UserID string `json:"userId"`
	// ResponsibleID stores the selected target responsible person id.
	ResponsibleID string `json:"responsibleId"`
}

// ConfirmTicketRequest stores rating (confirm_sc) payload: mark 0–5, comment required for 0–2.
type ConfirmTicketRequest struct {
	// UserID stores the acting MAX user identifier.
	UserID string `json:"userId"`
	// Mark stores the user rating 0–5 (legacy ITILIUM confirm_sc).
	Mark int `json:"mark"`
	// Comment stores optional explanation; required when Mark is 0, 1, or 2.
	Comment string `json:"comment"`
}

// ResponsibleOption stores one available responsible person.
type ResponsibleOption struct {
	// Team stores the team name.
	Team string `json:"team"`
	// Person stores the person display name.
	Person string `json:"person"`
	// Role stores the person's role.
	Role string `json:"role"`
	// ExternalID stores the ITILIUM identifier.
	ExternalID string `json:"externalId"`
}

// APIResponse provides a unified JSON envelope for handlers.
type APIResponse struct {
	// Success shows whether the operation finished successfully.
	Success bool `json:"success"`
	// Message stores a short explanation of the operation result.
	Message string `json:"message,omitempty"`
	// Data stores the payload returned to the client.
	Data any `json:"data,omitempty"`
	// RequestID stores the request correlation id for diagnostics.
	RequestID string `json:"requestId,omitempty"`
}
