// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package ruletable

import (
	"fmt"
	"strings"
)

// ErrScopeMismatch is returned when a request would combine a principal and a
// resource from different tenant scopes in a single authorization decision.
type ErrScopeMismatch struct {
	PrincipalID     string
	ResourceKind    string
	ResourceID      string
	PrincipalScope  string
	ResourceScope   string
	PrincipalTenant string
	ResourceTenant  string
}

func (e ErrScopeMismatch) Error() string {
	principal := fmt.Sprintf("principal scope %q", e.PrincipalScope)
	if e.PrincipalID != "" {
		principal = fmt.Sprintf("principal %q (scope %q)", e.PrincipalID, e.PrincipalScope)
	}

	resource := fmt.Sprintf("resource scope %q", e.ResourceScope)
	if e.ResourceKind != "" && e.ResourceID != "" {
		resource = fmt.Sprintf("resource kind %q with ID %q (scope %q)", e.ResourceKind, e.ResourceID, e.ResourceScope)
	} else if e.ResourceKind != "" {
		resource = fmt.Sprintf("resource kind %q (scope %q)", e.ResourceKind, e.ResourceScope)
	} else if e.ResourceID != "" {
		resource = fmt.Sprintf("resource %q (scope %q)", e.ResourceID, e.ResourceScope)
	}

	return fmt.Sprintf("%s and %s belong to different tenants (%q and %q); a request must not mix tenant scopes",
		principal, resource, e.PrincipalTenant, e.ResourceTenant)
}

// Is treats every ErrScopeMismatch as equivalent so callers can detect the
// condition regardless of the concrete identities involved.
func (e ErrScopeMismatch) Is(target error) bool {
	switch target.(type) {
	case ErrScopeMismatch, *ErrScopeMismatch:
		return true
	default:
		return false
	}
}

// As supports matching either an ErrScopeMismatch value or pointer so callers
// can use errors.As with the form that is most convenient for them.
func (e *ErrScopeMismatch) As(target any) bool {
	switch t := target.(type) {
	case *ErrScopeMismatch:
		*t = *e
		return true
	case **ErrScopeMismatch:
		*t = e
		return true
	default:
		return false
	}
}

// ErrTenantMismatch is an alias for ErrScopeMismatch. A tenant is represented
// by the root of a policy scope in Cerbos.
type ErrTenantMismatch = ErrScopeMismatch

// scopeTenant returns the tenant root of a scope. An empty scope has no tenant
// and is compatible with every tenant.
func scopeTenant(scope string) string {
	scope = strings.TrimPrefix(scope, ".")
	if scope == "" {
		return ""
	}

	tenant, _, _ := strings.Cut(scope, ".")
	return tenant
}

// validateScopeIsolation ensures that the principal and resource scopes of a
// request resolve to the same tenant. Scopes that are empty or share the same
// root are left alone, so requests within one tenant keep their existing
// behaviour. The error retains both identities so that the rejected decision
// can be attributed to the exact principal and resource that caused it.
func validateScopeIsolation(principalID, principalScope, resourceKind, resourceID, resourceScope string) error {
	principalTenant := scopeTenant(principalScope)
	resourceTenant := scopeTenant(resourceScope)

	if principalTenant == "" || resourceTenant == "" || principalTenant == resourceTenant {
		return nil
	}

	return &ErrScopeMismatch{
		PrincipalID:     principalID,
		ResourceKind:    resourceKind,
		ResourceID:      resourceID,
		PrincipalScope:  principalScope,
		ResourceScope:   resourceScope,
		PrincipalTenant: principalTenant,
		ResourceTenant:  resourceTenant,
	}
}

// ValidateScopeIsolation exposes the tenant-scope admission check to callers
// that evaluate requests without going through the rule table, such as the
// engine's request admission path.
func ValidateScopeIsolation(principalID, principalScope, resourceKind, resourceID, resourceScope string) error {
	return validateScopeIsolation(principalID, principalScope, resourceKind, resourceID, resourceScope)
}
