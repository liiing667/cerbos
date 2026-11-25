// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"testing"

	"github.com/stretchr/testify/require"

	enginev1 "github.com/cerbos/cerbos/api/genpb/cerbos/engine/v1"
	"github.com/cerbos/cerbos/internal/util"
)

func TestCheckRejectsAmbiguousActionsBeforeEvaluation(t *testing.T) {
	t.Parallel()

	auditLog := &mockAuditLog{}
	eng := newEngine(nil, Components{AuditLog: auditLog})

	outputs, err := eng.Check(t.Context(), []*enginev1.CheckInput{{
		Actions: []string{"view", " view"},
	}})

	require.ErrorIs(t, err, util.ErrInvalidAction)
	require.Nil(t, outputs)
	require.Empty(t, auditLog.decisionLogs, "rejected actions must not be written to the decision log")
}

func TestPlanRejectsAmbiguousActionsBeforeEvaluation(t *testing.T) {
	t.Parallel()

	auditLog := &mockAuditLog{}
	eng := newEngine(nil, Components{AuditLog: auditLog})

	output, err := eng.Plan(t.Context(), &enginev1.PlanResourcesInput{
		Actions: []string{"view", " view"},
	})

	require.ErrorIs(t, err, util.ErrInvalidAction)
	require.Nil(t, output)
	require.Empty(t, auditLog.decisionLogs, "rejected actions must not be written to the decision log")
}
