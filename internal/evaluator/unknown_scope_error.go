// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package evaluator

import (
	"errors"
	"fmt"
)

// ErrUnknownScope is the sentinel error for requests that reference a scope which is not present in the policy store.
var ErrUnknownScope = errors.New("unknown scope")

// ErrScopeNotFound is an alias for ErrUnknownScope.
var ErrScopeNotFound = ErrUnknownScope

// ErrMissingScope is an alias for ErrUnknownScope.
var ErrMissingScope = ErrUnknownScope

// UnknownScopeError is returned when a request references a non-empty scope that does not exist in the policy
// store and lenient scope search is disabled. It carries the requested scope and policy version so that callers can
// distinguish a missing scope from an ordinary policy decision.
type UnknownScopeError struct {
	Scope         string
	PolicyVersion string
	// Version is an alias for PolicyVersion, retained for callers that refer to the policy version simply as Version.
	Version string
}

// ScopeNotFoundError is an alias for UnknownScopeError.
type ScopeNotFoundError = UnknownScopeError

// MissingScopeError is an alias for UnknownScopeError.
type MissingScopeError = UnknownScopeError

func (e *UnknownScopeError) Error() string {
	version := e.PolicyVersion
	if version == "" {
		version = e.Version
	}

	return fmt.Sprintf("unknown scope %q for policy version %q", e.Scope, version)
}

func (e *UnknownScopeError) Unwrap() error {
	return ErrUnknownScope
}

func (e *UnknownScopeError) Is(target error) bool {
	_, ok := target.(*UnknownScopeError)
	return ok
}
