package api

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode"
)

// MarketingPermissionFromFindEmployeePayload merges ITILIUM find_employee payloads where boolean flags
// use different casing, Russian key names or nested wrappers (data/employee).
func MarketingPermissionFromFindEmployeePayload(payload map[string]any) bool {
	return permissionFromPayload(payload, [][]string{
		{"canCreateMarketingRequests", "canCreateMarketing", "can_marketing"},
		{"CanCreateMarketingRequests", "CanCreateMarketing"},
		{"СоздаватьЗаявкиМаркетинга", "СоздаватьМаркетинговыеЗаявки"},
		{"МожноСоздаватьЗаявкиМаркетинга", "МожноСоздаватьМаркетинговыеЗаявки"},
		{"РазрешитьСоздаватьЗаявкиМаркетинга", "РазрешитьМаркетинговыеЗаявки"},
	}, marketingKeyHeuristicMatch)
}

// DaxPermissionFromFindEmployeePayload merges DAX access flags returned by legacy find_employee JSON.
func DaxPermissionFromFindEmployeePayload(payload map[string]any) bool {
	return permissionFromPayload(payload, [][]string{
		{"canCreateDaxRequests", "canCreateDax", "can_dax"},
		{"CanCreateDaxRequests", "CanCreateDax"},
	}, daxKeyHeuristicMatch)
}

func marketingKeyHeuristicMatch(key string) bool {
	k := normalizeKeyTokens(key)
	switch {
	case strings.Contains(k, "маркет") && stringContainsAnyOf(k,
		"созда", "разреш", "можно", "заяв"):
		return true
	case strings.Contains(k, "marketing") && stringContainsAnyOf(k,
		"creat", "allow", "perm"):
		return true
	default:
		return false
	}
}

func daxKeyHeuristicMatch(key string) bool {
	k := normalizeKeyTokens(key)
	switch {
	case strings.Contains(k, "dax"):
		return true
	case strings.Contains(k, "дакс") && stringContainsAnyOf(k, "заяв", "созда", "разреш"):
		return true
	default:
		return false
	}
}

func stringContainsAnyOf(s string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

func permissionFromPayload(root map[string]any, synonyms [][]string, heuristic func(string) bool) bool {
	if root == nil {
		return false
	}

	for _, m := range unwrapPayloadMaps(root) {
		for _, group := range synonyms {
			if v := lookupBySynonyms(m, group); v != nil {
				if legacyBoolAny(v) {
					return true
				}
			}
		}

		for key, raw := range m {
			if heuristic(key) && legacyBoolAny(raw) {
				return true
			}
		}
	}

	return false
}

func unwrapPayloadMaps(root map[string]any) []map[string]any {
	out := []map[string]any{root}
	for _, envelope := range []string{"data", "Data", "result", "Result", "employee", "Employee"} {
		nested, ok := root[envelope].(map[string]any)
		if ok && nested != nil {
			out = append(out, nested)
		}
	}
	return out
}

func lookupBySynonyms(m map[string]any, synonyms []string) any {
	for _, target := range synonyms {
		want := normalizeKeyTokens(target)

		for key, val := range m {
			if normalizeKeyTokens(key) == want {
				return val
			}
		}
	}

	for _, target := range synonyms {
		for key, val := range m {
			if strings.EqualFold(key, target) {
				return val
			}
		}
	}

	return nil
}

func normalizeKeyTokens(key string) string {
	var b strings.Builder

	for _, r := range strings.TrimSpace(key) {
		switch r {
		case '_', '-', ' ', '"', '\'':
			continue
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}

	return b.String()
}

// legacyBoolAny accepts booleans plus string/number quirks from 1C JSON serializers.
func legacyBoolAny(value any) bool {
	switch converted := value.(type) {
	case bool:
		return converted
	case string:
		return boolFromRussifiedString(strings.TrimSpace(converted))
	case float64:
		return converted != 0
	case json.Number:
		return numericTruthy(string(converted))
	case int:
		return converted != 0
	case int64:
		return converted != 0
	default:
		return false
	}
}

func boolFromRussifiedString(s string) bool {
	if s == "" {
		return false
	}

	lower := strings.ToLower(s)

	switch lower {
	case "true", "1", "да", "y", "yes":
		return true
	case "false", "0", "нет", "no", "n":
		return false
	}

	if strings.Contains(lower, "истин") {
		return true
	}
	if strings.EqualFold(lower, "ложь") {
		return false
	}

	return false
}

func numericTruthy(raw string) bool {
	number, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)

	return err == nil && number != 0
}
