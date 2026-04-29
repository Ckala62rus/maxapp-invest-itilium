package api_test

import (
	"testing"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/api"
)

func TestMarketingPermissionFromFindEmployeePayload_synonyms(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		raw    map[string]any
		expect bool
	}{
		{name: "camelCase", raw: map[string]any{"canCreateMarketingRequests": true}, expect: true},
		{name: "PascalCase", raw: map[string]any{"CanCreateMarketingRequests": true}, expect: true},
		{name: "russian_truthy_string", raw: map[string]any{"СоздаватьМаркетинговыеЗаявки": "Истина"}, expect: true},
		{name: "numeric_one", raw: map[string]any{"можно_создавать_маркетинг": float64(1)}, expect: true},
		{name: "nested_data", raw: map[string]any{"data": map[string]any{"маркетинг_могу_создавать": true}}, expect: true},
		{name: "heuristic_key", raw: map[string]any{"разрешено_создание_заявок_маркетинга": true}, expect: true},
		{name: "false", raw: map[string]any{"canCreateMarketingRequests": false}, expect: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := api.MarketingPermissionFromFindEmployeePayload(tc.raw)
			if got != tc.expect {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestDaxPermissionFromFindEmployeePayload(t *testing.T) {
	t.Parallel()

	raw := map[string]any{"canCreateDaxRequests": true}

	if !api.DaxPermissionFromFindEmployeePayload(raw) {
		t.Fatalf("expected true for DAX synonym")
	}
}
