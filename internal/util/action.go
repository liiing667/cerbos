// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidAction is the sentinel error for an action identity that cannot be
// evaluated or published safely.
var ErrInvalidAction = errors.New("invalid action")

// InvalidActionError preserves the original blank action value.
//
// The value is quoted and its raw bytes are included so invisible whitespace is
// visible in errors and logs.
type InvalidActionError struct {
	Action string
}

func (e *InvalidActionError) Error() string {
	if e == nil {
		return ErrInvalidAction.Error()
	}

	return fmt.Sprintf("%s %q: action must not be blank (raw bytes: % X)",
		ErrInvalidAction, e.Action, []byte(e.Action))
}

func (*InvalidActionError) Unwrap() error {
	return ErrInvalidAction
}

// ValidateAction rejects an empty or whitespace-only action.
//
// Non-blank actions are returned to the caller unchanged: no trimming,
// case-folding, or wildcard normalization is performed.
func ValidateAction(action string) error {
	if strings.TrimSpace(action) == "" {
		return &InvalidActionError{Action: action}
	}

	return nil
}

// ValidateActions rejects any empty or whitespace-only action in actions.
// An empty action list is valid at this layer because some internal callers use
// it to represent a broader query.
func ValidateActions(actions []string) error {
	for _, action := range actions {
		if err := ValidateAction(action); err != nil {
			return err
		}
	}

	return nil
}
