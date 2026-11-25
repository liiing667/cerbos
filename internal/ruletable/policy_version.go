// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package ruletable

import "fmt"

// ErrPolicyVersionMismatch is returned when a request resolves the principal
// and resource policies to different policy versions. Policies from different
// versions must never be combined within a single authorization evaluation.
type ErrPolicyVersionMismatch struct {
	PrincipalVersion string
	ResourceVersion  string
}

func (e ErrPolicyVersionMismatch) Error() string {
	return fmt.Sprintf("principal policy version %q does not match resource policy version %q; a request must be evaluated against a single policy version",
		e.PrincipalVersion, e.ResourceVersion)
}

// Is treats every ErrPolicyVersionMismatch as equivalent so callers can detect
// the condition with errors.Is(err, ErrPolicyVersionMismatch{}) regardless of
// the concrete versions involved.
func (ErrPolicyVersionMismatch) Is(target error) bool {
	_, ok := target.(ErrPolicyVersionMismatch)
	return ok
}

// validatePolicyVersionIsolation ensures that principal and resource policies
// from different policy versions are never combined within a single
// authorization evaluation. A mismatch is only an error when policies of both
// kinds actually participate in the evaluation: if one kind has no matching
// policy, the request is decided by the other kind alone, as before.
func validatePolicyVersionIsolation(principalVersion, resourceVersion string, principalPolicyExists, resourcePolicyExists bool) error {
	if principalPolicyExists && resourcePolicyExists && principalVersion != resourceVersion {
		return ErrPolicyVersionMismatch{
			PrincipalVersion: principalVersion,
			ResourceVersion:  resourceVersion,
		}
	}

	return nil
}
