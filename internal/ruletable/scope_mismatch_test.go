// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package ruletable_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	effectv1 "github.com/cerbos/cerbos/api/genpb/cerbos/effect/v1"
	enginev1 "github.com/cerbos/cerbos/api/genpb/cerbos/engine/v1"
	policyv1 "github.com/cerbos/cerbos/api/genpb/cerbos/policy/v1"
	runtimev1 "github.com/cerbos/cerbos/api/genpb/cerbos/runtime/v1"
	"github.com/cerbos/cerbos/internal/conditions"
	"github.com/cerbos/cerbos/internal/engine/tracer"
	"github.com/cerbos/cerbos/internal/evaluator"
	"github.com/cerbos/cerbos/internal/ruletable"
	"github.com/cerbos/cerbos/internal/schema"
)

func scopeMismatchTestPrincipalPolicy(principal, version, scope string) *policyv1.Policy {
	return &policyv1.Policy{
		ApiVersion: "api.cerbos.dev/v1",
		PolicyType: &policyv1.Policy_PrincipalPolicy{
			PrincipalPolicy: &policyv1.PrincipalPolicy{
				Principal: principal,
				Version:   version,
				Scope:     scope,
				Rules: []*policyv1.PrincipalRule{{
					Resource: "document",
					Actions: []*policyv1.PrincipalRule_Action{{
						Action: "view",
						Effect: effectv1.Effect_EFFECT_ALLOW,
					}},
				}},
			},
		},
	}
}

func scopeMismatchTestEvalParams() evaluator.EvalParams {
	conf := &evaluator.Conf{}
	conf.SetDefaults()

	return evaluator.EvalParams{
		DefaultPolicyVersion: conf.DefaultPolicyVersion,
		DefaultScope:         conf.DefaultScope,
		NowFunc:              conditions.Now(),
	}
}

func TestScopeIsolation_Regression(t *testing.T) {
	principalRoot := scopeMismatchTestPrincipalPolicy("alice", "default", "")
	resourceRoot := scopedResourcePolicy("document", "", policyv1.ScopePermissions_SCOPE_PERMISSIONS_UNSPECIFIED)

	sets := []*runtimev1.RunnablePolicySet{
		compilePolicy(t, principalRoot),
		compilePolicy(t, scopeMismatchTestPrincipalPolicy("alice", "default", "tenant-a"), principalRoot),
		compilePolicy(t, scopeMismatchTestPrincipalPolicy("alice", "default", "tenant-b"), principalRoot),
		compilePolicy(t, resourceRoot),
		compilePolicy(t, scopedResourcePolicy("document", "tenant-a", policyv1.ScopePermissions_SCOPE_PERMISSIONS_UNSPECIFIED), resourceRoot),
		compilePolicy(t, scopedResourcePolicy("document", "tenant-b", policyv1.ScopePermissions_SCOPE_PERMISSIONS_UNSPECIFIED), resourceRoot),
	}

	loader := staticLoader{sets: sets}
	rt, err := ruletable.NewRuleTableFromLoader(t.Context(), loader)
	require.NoError(t, err)

	mgr, err := ruletable.NewRuleTableManager(rt, loader, schema.NewNopManager())
	require.NoError(t, err)

	params := scopeMismatchTestEvalParams()

	input := func(principalScope, resourceScope string) *enginev1.CheckInput {
		return &enginev1.CheckInput{
			RequestId: "scope-isolation",
			Actions:   []string{"view"},
			Principal: &enginev1.Principal{Id: "alice", Roles: []string{"user"}, Scope: principalScope},
			Resource:  &enginev1.Resource{Kind: "document", Id: "doc-1", Scope: resourceScope},
		}
	}

	t.Run("same tenant keeps the existing result", func(t *testing.T) {
		out, _, err := mgr.Check(t.Context(), tracer.Start(nil), params, input("tenant-a", "tenant-a"))
		require.NoError(t, err)
		require.Equal(t, effectv1.Effect_EFFECT_ALLOW, out.Actions["view"].GetEffect())
	})

	t.Run("unscoped side keeps the existing result", func(t *testing.T) {
		out, _, err := mgr.Check(t.Context(), tracer.Start(nil), params, input("", "tenant-a"))
		require.NoError(t, err)
		require.Equal(t, effectv1.Effect_EFFECT_ALLOW, out.Actions["view"].GetEffect())
	})

	t.Run("mixed tenants fail check before evaluation", func(t *testing.T) {
		out, _, err := mgr.Check(t.Context(), tracer.Start(nil), params, input("tenant-b", "tenant-a"))
		require.Nil(t, out)
		require.ErrorIs(t, err, ruletable.ErrScopeMismatch{})

		var mismatch ruletable.ErrScopeMismatch
		require.ErrorAs(t, err, &mismatch)
		require.Equal(t, "alice", mismatch.PrincipalID)
		require.Equal(t, "doc-1", mismatch.ResourceID)
		require.Equal(t, "tenant-b", mismatch.PrincipalScope)
		require.Equal(t, "tenant-a", mismatch.ResourceScope)

		msg := err.Error()
		require.True(t, strings.Contains(msg, "tenant-a"))
		require.True(t, strings.Contains(msg, "tenant-b"))
	})

	t.Run("mixed tenants fail plan before evaluation", func(t *testing.T) {
		out, _, err := mgr.Plan(t.Context(), params, &enginev1.PlanResourcesInput{
			RequestId: "scope-isolation-plan",
			Actions:   []string{"view"},
			Principal: &enginev1.Principal{Id: "alice", Roles: []string{"user"}, Scope: "tenant-b"},
			Resource:  &enginev1.PlanResourcesInput_Resource{Kind: "document", Scope: "tenant-a"},
		})
		require.Nil(t, out)
		require.ErrorIs(t, err, ruletable.ErrScopeMismatch{})
	})
}
