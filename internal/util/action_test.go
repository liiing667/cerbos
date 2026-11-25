// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cerbos/cerbos/internal/util"
)

func TestValidateActionsRejectsBlankActions(t *testing.T) {
	t.Parallel()

	for _, action := range []string{"", " ", "\t", "\n", " \t\n "} {
		err := util.ValidateAction(action)
		require.ErrorIs(t, err, util.ErrInvalidAction)

		var invalidErr *util.InvalidActionError
		require.ErrorAs(t, err, &invalidErr)
		require.Equal(t, action, invalidErr.Action)
	}

	err := util.ValidateActions([]string{"view", ""})
	require.ErrorIs(t, err, util.ErrInvalidAction)
}

func TestValidateActionsPreservesNormalAndWildcardActions(t *testing.T) {
	t.Parallel()

	actions := []string{"view", "edit details", "view:*", "*"}
	require.NoError(t, util.ValidateActions(actions))
}

func TestValidateActionsAllowsEmptySetForInternalCallers(t *testing.T) {
	t.Parallel()

	require.NoError(t, util.ValidateActions(nil))
	require.NoError(t, util.ValidateActions([]string{}))
}
