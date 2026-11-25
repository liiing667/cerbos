// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package index

import "strings"

// NormalizeRoles removes blank and duplicate role identities while preserving
// the order of the first occurrence of every distinct role.
//
// Surrounding whitespace is intentionally not trimmed from a non-blank value:
// callers and audit records retain the original identity, so silently changing
// it would make the evaluated role differ from the recorded one. Values that
// contain only whitespace carry no identity and are omitted.
func NormalizeRoles(roles []string) []string {
	if len(roles) == 0 {
		return roles
	}
	if len(roles) == 1 {
		if isBlankRole(roles[0]) {
			return []string{}
		}
		return roles
	}

	seen := make(map[string]struct{}, len(roles))
	changed := false
	for _, role := range roles {
		if strings.TrimSpace(role) == "" {
			changed = true
			break
		}
		if _, ok := seen[role]; ok {
			changed = true
			break
		}
		seen[role] = struct{}{}
	}
	if !changed {
		return roles
	}

	out := make([]string, 0, len(roles))
	seen = make(map[string]struct{}, len(roles))
	for _, role := range roles {
		if isBlankRole(role) {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	return out
}

func isBlankRole(role string) bool {
	return strings.TrimSpace(role) == ""
}

func normalizeIndexRole(role string) string {
	if isBlankRole(role) {
		return ""
	}
	return role
}
