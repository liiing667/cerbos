// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	enginev1 "github.com/cerbos/cerbos/api/genpb/cerbos/engine/v1"
	"github.com/cerbos/cerbos/internal/util"
)

func TestEngineRejectsBidirectionalPrincipalID(t *testing.T) {
	eng := NewEphemeral(nil, nil, nil)
	principal := &enginev1.Principal{Id: "alice\u202e"}

	_, err := eng.Check(context.Background(), []*enginev1.CheckInput{{Principal: principal}})
	require.ErrorIs(t, err, util.ErrInvalidPrincipalID)

	var invalidErr *util.InvalidPrincipalIDError
	require.ErrorAs(t, err, &invalidErr)
	require.Equal(t, principal.Id, invalidErr.PrincipalID)
	require.Contains(t, err.Error(), `alice\u202e`)

	_, err = eng.Plan(context.Background(), &enginev1.PlanResourcesInput{Principal: principal})
	require.ErrorIs(t, err, util.ErrInvalidPrincipalID)
	require.Contains(t, err.Error(), `alice\u202e`)
}
