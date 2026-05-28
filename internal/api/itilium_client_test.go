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
