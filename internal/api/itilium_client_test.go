package api

import (
	"strings"
	"testing"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
)

func TestBuildCreateSCLongDescriptionDoesNotAppendMetadata(t *testing.T) {
	t.Parallel()

	description := buildCreateSCLongDescription(models.CreateTicketRequest{
		RequestType:   "ИТ",
		Description:   "Нужно настроить доступ",
		Department:    "Бухгалтерия",
		ExecutionDate: "31.05.2026",
	})

	if description != "Нужно настроить доступ" {
		t.Fatalf("description = %q, want only user text", description)
	}
	for _, forbidden := range []string{"Тип:", "Подразделение:", "Исполнить до:"} {
		if strings.Contains(description, forbidden) {
			t.Fatalf("description must not contain %q: %q", forbidden, description)
		}
	}
}

func TestFormatMarketingExecutionDateConvertsHTMLDate(t *testing.T) {
	t.Parallel()

	got := formatMarketingExecutionDate("2026-06-26")

	if got != "2026-06-26" {
		t.Fatalf("formatMarketingExecutionDate() = %q, want %q", got, "2026-06-26")
	}
}

func TestFormatMarketingExecutionDateConvertsDottedDate(t *testing.T) {
	t.Parallel()

	got := formatMarketingExecutionDate("26.06.2026")

	if got != "2026-06-26" {
		t.Fatalf("formatMarketingExecutionDate() = %q, want %q", got, "2026-06-26")
	}
}

func TestParseCreateSCResponseReturnsStringBusinessError(t *testing.T) {
	t.Parallel()

	_, err := parseCreateSCResponse(
		[]byte(`"Не указана услуга. Не указано подразделение. Необходимо указать желаемую дату исполнения."`),
		models.CreateTicketRequest{Title: "Маркетинговая заявка"},
	)

	if err == nil {
		t.Fatal("parseCreateSCResponse() error = nil, want business error")
	}
	if !strings.Contains(err.Error(), "Не указана услуга") {
		t.Fatalf("parseCreateSCResponse() error = %q, want business message", err.Error())
	}
}

func TestIsMarketingRequiredFieldsMissingResponse(t *testing.T) {
	t.Parallel()

	if !isMarketingRequiredFieldsMissingResponse([]byte(`"Не указана услуга. Не указано подразделение. Необходимо указать желаемую дату исполнения."`)) {
		t.Fatal("isMarketingRequiredFieldsMissingResponse() = false, want true")
	}
	if isMarketingRequiredFieldsMissingResponse([]byte(`{"number":"0000000001"}`)) {
		t.Fatal("isMarketingRequiredFieldsMissingResponse() = true for success payload")
	}
}

func TestIsMarketingExecutionDateOnlyMissingResponse(t *testing.T) {
	t.Parallel()

	if !isMarketingExecutionDateOnlyMissingResponse([]byte(`" Необходимо указать желаемую дату исполнения."`)) {
		t.Fatal("isMarketingExecutionDateOnlyMissingResponse() = false, want true for date-only error")
	}
	if isMarketingExecutionDateOnlyMissingResponse([]byte(`"Не указана услуга. Необходимо указать желаемую дату исполнения."`)) {
		t.Fatal("isMarketingExecutionDateOnlyMissingResponse() = true, want false when service is also missing")
	}
}

func TestSplitMarketingCreateFormSeparatesDateFields(t *testing.T) {
	t.Parallel()

	form := buildMarketingCreateForm(models.CreateMarketingRequest{
		UserID:        "40367639",
		ServiceCode:   "SMM",
		FormNumber:    "3",
		Subdivision:   "Иван Васильевич",
		ExecutionDate: "2026-07-08",
		FormData:      map[string]string{"Description": "test"},
	})

	query, dateBody := splitMarketingCreateForm(form)
	if query.Get("Services") != "SMM" {
		t.Fatalf("query Services = %q, want SMM", query.Get("Services"))
	}
	if query.Get("WithoutDate") != "Ложь" {
		t.Fatalf("query WithoutDate = %q, want Ложь", query.Get("WithoutDate"))
	}
	if query.Get("ExecutionDate") != "" {
		t.Fatalf("query must not contain ExecutionDate, got %q", query.Get("ExecutionDate"))
	}
	if dateBody.Get("ExecutionDate") != "2026-07-08" {
		t.Fatalf("dateBody ExecutionDate = %q, want 2026-07-08", dateBody.Get("ExecutionDate"))
	}
	if len(dateBody) != 1 {
		t.Fatalf("dateBody must contain only ExecutionDate, got %v", dateBody)
	}
}

func TestMarketingDateFieldVariantsIncludeDocumentedExecutionDate(t *testing.T) {
	t.Parallel()

	variants := marketingDateFieldVariants("2026-07-01")
	if len(variants) == 0 {
		t.Fatal("marketingDateFieldVariants() returned no variants")
	}
	if variants[0].Get("ExecutionDate") != "2026-07-01" {
		t.Fatalf("first variant ExecutionDate = %q, want 2026-07-01", variants[0].Get("ExecutionDate"))
	}
}
