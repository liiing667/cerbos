// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cerbos/cerbos/internal/util"
)

func TestValidateActions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		actions []string
		wantErr bool
	}{
		{
			name:    "valid actions preserve set and order",
			actions: []string{"view", "edit", "delete", "view:*", "*"},
		},
		{
			name:    "case is significant",
			actions: []string{"View", "view"},
		},
		{
			name:    "action with internal whitespace",
			actions: []string{"view details"},
		},
		{
			name:    "empty list",
			actions: nil,
		},
		{
			name:    "leading whitespace",
			actions: []string{" view"},
			wantErr: true,
		},
		{
			name:    "trailing whitespace",
			actions: []string{"view "},
			wantErr: true,
		},
		{
			name:    "blank action",
			actions: []string{" \t\n"},
			wantErr: true,
		},
		{
			name:    "duplicate action",
			actions: []string{"view", "view"},
			wantErr: true,
		},
		{
			name:    "whitespace variant of an earlier action",
			actions: []string{"view", " view"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := util.ValidateActions(tc.actions)
			if !tc.wantErr {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, util.ErrInvalidAction)
		})
	}
}

func TestInvalidActionErrorPreservesOriginalValue(t *testing.T) {
	t.Parallel()

	err := util.ValidateActions([]string{"view", " view"})
	require.ErrorIs(t, err, util.ErrInvalidAction)

	var invalidErr *util.InvalidActionError
	require.ErrorAs(t, err, &invalidErr)
	require.Equal(t, " view", invalidErr.Action)
	require.Equal(t, "view", invalidErr.Conflict)
	require.Contains(t, err.Error(), `" view"`)
	require.Contains(t, err.Error(), `"view"`)
	require.Contains(t, err.Error(), "20 76 69 65 77")
}

func TestValidateAction(t *testing.T) {
	t.Parallel()

	require.NoError(t, util.ValidateAction("view"))
	require.ErrorIs(t, util.ValidateAction(" "), util.ErrInvalidAction)
}
