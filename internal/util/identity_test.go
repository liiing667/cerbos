// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cerbos/cerbos/internal/util"
)

func TestValidatePrincipalID(t *testing.T) {
	validIDs := []string{
		"alice",
		"山田太郎",
		"مستخدم", // Arabic text without explicit bidi controls
		"emily jones",
		"e\u0301lodie",
		"अनुजा",
	}

	for _, id := range validIDs {
		t.Run("valid/"+id, func(t *testing.T) {
			require.NoError(t, util.ValidatePrincipalID(id))
		})
	}

	controlIDs := map[rune]string{
		'\u061c': "Arabic letter mark",
		'\u200e': "left-to-right mark",
		'\u200f': "right-to-left mark",
		'\u202a': "left-to-right embedding",
		'\u202b': "right-to-left embedding",
		'\u202c': "pop directional formatting",
		'\u202d': "left-to-right override",
		'\u202e': "right-to-left override",
		'\u2066': "left-to-right isolate",
		'\u2067': "right-to-left isolate",
		'\u2068': "first strong isolate",
		'\u2069': "pop directional isolate",
	}

	for r, name := range controlIDs {
		id := "alice" + string(r)
		t.Run("invalid/"+name, func(t *testing.T) {
			err := util.ValidatePrincipalID(id)
			require.Error(t, err)
			require.ErrorIs(t, err, util.ErrInvalidPrincipalID)

			var invalidErr *util.InvalidPrincipalIDError
			require.ErrorAs(t, err, &invalidErr)
			require.Equal(t, id, invalidErr.PrincipalID)

			msg := err.Error()
			require.Contains(t, msg, strconv.QuoteToASCII(id))
			require.Contains(t, msg, fmt.Sprintf("% X", []byte(id)))
			require.Contains(t, msg, "raw bytes:")
		})
	}
}
