// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"fmt"
	"unicode"
)

// ErrInvalidPrincipalID is the sentinel error for an unsafe principal ID.
var ErrInvalidPrincipalID = errors.New("invalid principal ID")

// InvalidPrincipalIDError preserves the original principal ID and its bytes.
//
// The ID is quoted with %q when rendered so invisible formatting runes become
// visible escape sequences instead of changing the direction of the log or
// error message.
type InvalidPrincipalIDError struct {
	PrincipalID string
}

func (e *InvalidPrincipalIDError) Error() string {
	if e == nil {
		return ErrInvalidPrincipalID.Error()
	}

	return fmt.Sprintf("%s %q: contains bidirectional control characters (raw bytes: % X)",
		ErrInvalidPrincipalID, e.PrincipalID, []byte(e.PrincipalID))
}

func (*InvalidPrincipalIDError) Unwrap() error {
	return ErrInvalidPrincipalID
}

// ValidatePrincipalID rejects principal IDs containing Unicode bidirectional
// control characters. Ordinary Unicode letters, marks and RTL languages
// without explicit bidi controls remain valid.
func ValidatePrincipalID(id string) error {
	for _, r := range id {
		if unicode.In(r, unicode.Bidi_Control) {
			return &InvalidPrincipalIDError{PrincipalID: id}
		}
	}

	return nil
}
