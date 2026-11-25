// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidAction is the sentinel error for an action identity that cannot be
// used unambiguously.
var ErrInvalidAction = errors.New("invalid action")

// InvalidActionError preserves the original action value and, when the action
// collides with another action after surrounding whitespace is ignored, the
// action it conflicts with.
//
// Values are quoted with %q and their raw bytes are included so that leading,
// trailing, and invisible whitespace remains visible in errors and logs.
type InvalidActionError struct {
	Action   string
	Conflict string
}

func (e *InvalidActionError) Error() string {
	if e == nil {
		return ErrInvalidAction.Error()
	}

	if e.Conflict != "" {
		return fmt.Sprintf("%s %q: conflicts with action %q (raw bytes: % X)",
			ErrInvalidAction, e.Action, e.Conflict, []byte(e.Action))
	}

	return fmt.Sprintf("%s %q: action must not be blank or have leading or trailing whitespace (raw bytes: % X)",
		ErrInvalidAction, e.Action, []byte(e.Action))
}

func (*InvalidActionError) Unwrap() error {
	return ErrInvalidAction
}

// ValidateAction validates a single action identity.
func ValidateAction(action string) error {
	return ValidateActions([]string{action})
}

// ValidateActions rejects action identities that are blank, have leading or
// trailing whitespace, or collide with another action once surrounding
// whitespace is ignored (including exact duplicates).
//
// Valid actions are never trimmed, deduplicated, or reordered: the caller's
// action set and order are preserved so that policy evaluation and audit
// records keep the original identity. Actions that differ only by case remain
// distinct, and wildcard actions such as "*" and "view:*" are unaffected.
func ValidateActions(actions []string) error {
	if len(actions) == 0 {
		return nil
	}

	// Keyed by the whitespace-insensitive form of each action. The value is the
	// original action that established that identity.
	seen := make(map[string]string, len(actions))
	for _, action := range actions {
		identity := strings.TrimSpace(action)
		if identity == "" {
			return &InvalidActionError{Action: action}
		}

		if previous, ok := seen[identity]; ok {
			return &InvalidActionError{Action: action, Conflict: previous}
		}

		seen[identity] = action
		if action != identity {
			return &InvalidActionError{Action: action}
		}
	}

	return nil
}
