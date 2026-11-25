// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package ruletable_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	effectv1 "github.com/cerbos/cerbos/api/genpb/cerbos/effect/v1"
	enginev1 "github.com/cerbos/cerbos/api/genpb/cerbos/engine/v1"
	policyv1 "github.com/cerbos/cerbos/api/genpb/cerbos/policy/v1"
	runtimev1 "github.com/cerbos/cerbos/api/genpb/cerbos/runtime/v1"
	"github.com/cerbos/cerbos/internal/evaluator"
	"github.com/cerbos/cerbos/internal/ruletable"
	"github.com/cerbos/cerbos/internal/schema"
)

func TestUnknownScopeResolutionPreservesKnownEmptyScopesAndInheritance(t *testing.T) {
	root := scopedResourcePolicy("document", "", unspecifiedPerm)
	emptyTenant := &policyv1.Policy{
		ApiVersion: "api.cerbos.dev/v1",
		PolicyType: &policyv1.Policy_ResourcePolicy{
			ResourcePolicy: &policyv1.ResourcePolicy{
				Resource: "document",
				Version:  "default",
				Scope:    "tenant",
			},
		},
	}

	sets := []*runtimev1.RunnablePolicySet{
		compilePolicy(t, root),
		compilePolicy(t, emptyTenant, root),
	}

	rt, err := ruletable.NewRuleTableFromLoader(t.Context(), staticLoader{sets: sets})
	require.NoError(t, err)

	schemaMgr := schema.NewNopManager()
	conf := &evaluator.Conf{}
	conf.SetDefaults()

	checkInput := func(scope string) *enginev1.CheckInput {
		return &enginev1.CheckInput{
			RequestId: "test",
			Actions:   []string{"view"},
			Principal: &enginev1.Principal{Id: "sam", Roles: []string{"user"}},
			Resource:  &enginev1.Resource{Kind: "document", Id: "1", Scope: scope},
		}
	}

	planInput := func(scope string) *enginev1.PlanResourcesInput {
		return &enginev1.PlanResourcesInput{
			RequestId: "test",
			Actions:   []string{"view"},
			Principal: &enginev1.Principal{Id: "sam", Roles: []string{"user"}},
			Resource:  &enginev1.PlanResourcesInput_Resource{Kind: "document", Scope: scope},
		}
	}

	t.Run("check_fails_for_unknown_scope", func(t *testing.T) {
		_, _, err := rt.Check(t.Context(), conf, schemaMgr, []*enginev1.CheckInput{checkInput("tenant.unknown")})
		requireUnknownScopeFailure(t, err, "tenant.unknown", "default")
	})

	t.Run("plan_fails_for_unknown_scope", func(t *testing.T) {
		_, _, err := rt.Plan(t.Context(), conf, schemaMgr, planInput("tenant.unknown"))
		requireUnknownScopeFailure(t, err, "tenant.unknown", "default")
	})

	t.Run("lenient_search_still_falls_back_to_a_parent", func(t *testing.T) {
		lenientConf := *conf
		lenientConf.LenientScopeSearch = true

		out, _, err := rt.Check(t.Context(), &lenientConf, schemaMgr, []*enginev1.CheckInput{checkInput("tenant.unknown")})
		require.NoError(t, err)
		require.Equal(t, effectv1.Effect_EFFECT_ALLOW, out[0].Actions["view"].GetEffect())
	})

	t.Run("known_empty_scope_inherits_from_parent", func(t *testing.T) {
		out, _, err := rt.Check(t.Context(), conf, schemaMgr, []*enginev1.CheckInput{checkInput("tenant")})
		require.NoError(t, err)
		require.Equal(t, effectv1.Effect_EFFECT_ALLOW, out[0].Actions["view"].GetEffect())
	})
}

func requireUnknownScopeFailure(t *testing.T, err error, scope, version string) {
	t.Helper()

	require.Error(t, err)
	require.ErrorIs(t, err, ruletable.ErrUnknownScope)

	var unknownScopeErr *ruletable.UnknownScopeError
	require.True(t, errors.As(err, &unknownScopeErr))
	require.Equal(t, scope, unknownScopeErr.Scope)
	require.Equal(t, version, unknownScopeErr.PolicyVersion)
	require.Contains(t, err.Error(), scope)
	require.Contains(t, err.Error(), version)
}
