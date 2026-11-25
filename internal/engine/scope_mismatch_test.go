// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"testing"

	"github.com/stretchr/testify/require"

	enginev1 "github.com/cerbos/cerbos/api/genpb/cerbos/engine/v1"
	"github.com/cerbos/cerbos/internal/ruletable"
)

func TestScopeIsolationAtAdmission_Regression(t *testing.T) {
	eng := NewEphemeral(nil, nil, nil)

	t.Run("check", func(t *testing.T) {
		_, err := eng.Check(t.Context(), []*enginev1.CheckInput{{
			RequestId: "scope-admission",
			Actions:   []string{"view"},
			Principal: &enginev1.Principal{Id: "alice", Roles: []string{"user"}, Scope: "tenant-b"},
			Resource:  &enginev1.Resource{Kind: "document", Id: "doc-1", Scope: "tenant-a"},
		}})
		require.ErrorIs(t, err, ruletable.ErrScopeMismatch{})
	})

	t.Run("plan", func(t *testing.T) {
		_, err := eng.Plan(t.Context(), &enginev1.PlanResourcesInput{
			RequestId: "scope-admission-plan",
			Actions:   []string{"view"},
			Principal: &enginev1.Principal{Id: "alice", Roles: []string{"user"}, Scope: "tenant-b"},
			Resource:  &enginev1.PlanResourcesInput_Resource{Kind: "document", Scope: "tenant-a"},
		})
		require.ErrorIs(t, err, ruletable.ErrScopeMismatch{})
	})
}
