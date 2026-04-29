// Package api_test проверяет публичные функции клиента ITILIUM/1С.
// Эти тесты лежат отдельно от internal/api, но могут импортировать internal-пакет,
// потому что директория tests находится внутри того же Go-модуля.
package api_test

import (
	"testing"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/api"
)

func TestMarketingPermissionFromFindEmployeePayload_synonyms(t *testing.T) {
	// Разрешаем параллельный запуск: тесты только читают локальные map и не меняют общее состояние.
	t.Parallel()

	// Это table-driven test: один и тот же алгоритм проверяется набором кейсов.
	// Такой стиль удобен, когда у функции много входных вариантов.
	cases := []struct {
		// name попадёт в имя подтеста: go test -run '.../camelCase'.
		name string
		// raw имитирует сырой JSON-объект, который пришёл от find_employee в 1С.
		raw map[string]any
		// expect — ожидаемый результат функции.
		expect bool
	}{
		// Английское camelCase поле — самый ожидаемый вариант.
		{name: "camelCase", raw: map[string]any{"canCreateMarketingRequests": true}, expect: true},
		// PascalCase нужен на случай, если 1С/JSON-сериализатор меняет регистр ключей.
		{name: "PascalCase", raw: map[string]any{"CanCreateMarketingRequests": true}, expect: true},
		// 1С часто отдаёт булевы значения строкой "Истина".
		{name: "russian_truthy_string", raw: map[string]any{"СоздаватьМаркетинговыеЗаявки": "Истина"}, expect: true},
		// Некоторые интеграции кодируют true числом 1.
		{name: "numeric_one", raw: map[string]any{"можно_создавать_маркетинг": float64(1)}, expect: true},
		// Иногда полезные поля лежат внутри вложенной обёртки data.
		{name: "nested_data", raw: map[string]any{"data": map[string]any{"маркетинг_могу_создавать": true}}, expect: true},
		// Эвристика ловит ключи с "маркетинг" + "создание/разрешено".
		{name: "heuristic_key", raw: map[string]any{"разрешено_создание_заявок_маркетинга": true}, expect: true},
		// Явный false не должен случайно превращаться в true.
		{name: "false", raw: map[string]any{"canCreateMarketingRequests": false}, expect: false},
	}

	// range проходит по всем кейсам.
	for _, tc := range cases {
		// В Go до версии 1.22 это защищало от захвата переменной цикла в parallel subtest.
		// Даже сейчас строка остаётся понятной и безопасной.
		tc := tc
		// t.Run создаёт отдельный подтест с именем tc.name.
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act: вызываем функцию, которую проверяем.
			got := api.MarketingPermissionFromFindEmployeePayload(tc.raw)
			// Assert: сравниваем фактический результат с ожидаемым.
			if got != tc.expect {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestDaxPermissionFromFindEmployeePayload(t *testing.T) {
	t.Parallel()

	// Минимальный happy path: поле DAX пришло в ожидаемом camelCase.
	raw := map[string]any{"canCreateDaxRequests": true}

	// Проверяем через if, чтобы было видно стандартный способ без testify.
	if !api.DaxPermissionFromFindEmployeePayload(raw) {
		t.Fatalf("expected true for DAX synonym")
	}
}
