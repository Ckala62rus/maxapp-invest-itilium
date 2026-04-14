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
	// Department stores the current department or branch.
	Department string `json:"department"`
	// EmployeeFound shows whether ITILIUM already knows this user.
	EmployeeFound bool `json:"employeeFound"`
	// RegistrationRequired tells the UI to open the registration form.
	RegistrationRequired bool `json:"registrationRequired"`
}

// RegistrationRequest describes the form used when the user is missing in ITILIUM.
type RegistrationRequest struct {
	// UserID stores the external MAX user identifier.
	UserID string `json:"userId"`
	// EmployeeNumber stores the user's internal employee number.
	EmployeeNumber string `json:"employeeNumber"`
	// FullName stores the user's full legal name.
	FullName string `json:"fullName"`
	// Department stores the user's store, branch or department.
	Department string `json:"department"`
	// Phone stores the callback phone number.
	Phone string `json:"phone"`
	// Comment stores free-form extra context.
	Comment string `json:"comment"`
}

// EmployeeLookupRequest describes a legacy-style user lookup in ITILIUM.
type EmployeeLookupRequest struct {
	// Identifier stores the value sent to ITILIUM, for example employee id or MAX user id.
	Identifier string `json:"identifier"`
	// AttributeCode stores the ITILIUM field name used for lookup, for example employee or telegram.
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
	// State stores the current workflow state.
	State string `json:"state"`
	// Deadline stores the current SLA or requested date.
	Deadline string `json:"deadline"`
	// ResponsibleTeam stores the team currently assigned to the ticket.
	ResponsibleTeam string `json:"responsibleTeam"`
	// CanChangeResponsible shows whether responsible reassignment is allowed.
	CanChangeResponsible bool `json:"canChangeResponsible"`
	// AvailableStates lists next state transition names.
	AvailableStates []string `json:"availableStates"`
	// Timeline stores comments and important system events.
	Timeline []CommentEntry `json:"timeline"`
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
