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

	if got != "26.06.2026" {
		t.Fatalf("formatMarketingExecutionDate() = %q, want %q", got, "26.06.2026")
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
