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

func TestCheckRejectsBlankActionBeforePublishingState(t *testing.T) {
	t.Parallel()

	auditLog := &mockAuditLog{}
	eng := newEngine(nil, Components{AuditLog: auditLog})

	outputs, err := eng.Check(context.Background(), []*enginev1.CheckInput{{
		Actions: []string{"view", " "},
	}})

	require.ErrorIs(t, err, util.ErrInvalidAction)
	require.Nil(t, outputs)
	require.Empty(t, auditLog.decisionLogs, "rejected action must not publish a decision")
}
