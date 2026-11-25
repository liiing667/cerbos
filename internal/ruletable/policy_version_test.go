// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package ruletable_test

import (
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

const (
	policyVersionDefault = "default"
	policyVersionBeta    = "beta"
)

func versionedResourcePolicy(resource, version string) *policyv1.Policy {
	return &policyv1.Policy{
		ApiVersion: "api.cerbos.dev/v1",
		PolicyType: &policyv1.Policy_ResourcePolicy{
			ResourcePolicy: &policyv1.ResourcePolicy{
				Resource: resource,
				Version:  version,
				Rules: []*policyv1.ResourceRule{
					{
						Actions: []string{"view"},
						Roles:   []string{"user"},
						Effect:  effectv1.Effect_EFFECT_ALLOW,
					},
				},
			},
		},
	}
}

func versionedPrincipalPolicy(principal, version string) *policyv1.Policy {
	return &policyv1.Policy{
		ApiVersion: "api.cerbos.dev/v1",
		PolicyType: &policyv1.Policy_PrincipalPolicy{
			PrincipalPolicy: &policyv1.PrincipalPolicy{
				Principal: principal,
				Version:   version,
				Rules: []*policyv1.PrincipalRule{
					{
						Resource: "document",
						Actions: []*policyv1.PrincipalRule_Action{
							{Action: "view", Effect: effectv1.Effect_EFFECT_ALLOW},
						},
					},
				},
			},
		},
	}
}

func versionEvalParams() evaluator.EvalParams {
	conf := &evaluator.Conf{}
	conf.SetDefaults()

	return evaluator.EvalParams{
		DefaultPolicyVersion: conf.DefaultPolicyVersion,
		DefaultScope:         conf.DefaultScope,
		NowFunc:              conditions.Now(),
	}
}

// A request must not combine principal and resource policies from different
// policy versions: mixed policy versions were previously evaluated together and
// could produce an ALLOW decision assembled across policy generations.
func TestPolicyVersionIsolation(t *testing.T) {
	const (
		action      = "view"
		resource    = "document"
		principalID = "alice"
	)

	resourceDefault := compilePolicy(t, versionedResourcePolicy(resource, policyVersionDefault))
	resourceBeta := compilePolicy(t, versionedResourcePolicy(resource, policyVersionBeta))
	principalDefault := compilePolicy(t, versionedPrincipalPolicy(principalID, policyVersionDefault))
	principalBeta := compilePolicy(t, versionedPrincipalPolicy(principalID, policyVersionBeta))

	loader := staticLoader{sets: []*runtimev1.RunnablePolicySet{
		resourceDefault, resourceBeta, principalDefault, principalBeta,
	}}

	rt, err := ruletable.NewRuleTableFromLoader(t.Context(), loader)
	require.NoError(t, err)

	mgr, err := ruletable.NewRuleTableManager(rt, loader, schema.NewNopManager())
	require.NoError(t, err)

	params := versionEvalParams()
	principal := func(version string) *enginev1.Principal {
		return &enginev1.Principal{Id: principalID, Roles: []string{"user"}, PolicyVersion: version}
	}

	t.Run("consistent_versions_keep_existing_result", func(t *testing.T) {
		for _, version := range []string{policyVersionDefault, policyVersionBeta} {
			out, _, err := mgr.Check(t.Context(), tracer.Start(nil), params, &enginev1.CheckInput{
				RequestId: "consistent-" + version,
				Resource:  &enginev1.Resource{Kind: resource, Id: "1", PolicyVersion: version},
				Principal: principal(version),
				Actions:   []string{action},
			})
			require.NoError(t, err)
			require.Equal(t, effectv1.Effect_EFFECT_ALLOW, out.Actions[action].GetEffect())
		}
	})

	t.Run("mixed_versions_fail_check", func(t *testing.T) {
		cases := []struct {
			name             string
			principalVersion string
			resourceVersion  string
		}{
			{
				name:             "principal_beta_resource_default",
				principalVersion: policyVersionBeta,
				resourceVersion:  policyVersionDefault,
			},
			{
				name:             "principal_default_resource_beta",
				principalVersion: policyVersionDefault,
				resourceVersion:  policyVersionBeta,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := mgr.Check(t.Context(), tracer.Start(nil), params, &enginev1.CheckInput{
					RequestId: "mixed-" + tc.name,
					Resource:  &enginev1.Resource{Kind: resource, Id: "1", PolicyVersion: tc.resourceVersion},
					Principal: principal(tc.principalVersion),
					Actions:   []string{action},
				})
				require.ErrorIs(t, err, ruletable.ErrPolicyVersionMismatch{})
			})
		}
	})

	t.Run("mixed_versions_fail_plan", func(t *testing.T) {
		cases := []struct {
			name             string
			principalVersion string
			resourceVersion  string
		}{
			{
				name:             "principal_beta_resource_default",
				principalVersion: policyVersionBeta,
				resourceVersion:  policyVersionDefault,
			},
			{
				name:             "principal_default_resource_beta",
				principalVersion: policyVersionDefault,
				resourceVersion:  policyVersionBeta,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := mgr.Plan(t.Context(), params, &enginev1.PlanResourcesInput{
					RequestId: "mixed-plan-" + tc.name,
					Actions:   []string{action},
					Resource:  &enginev1.PlanResourcesInput_Resource{Kind: resource, PolicyVersion: tc.resourceVersion},
					Principal: principal(tc.principalVersion),
				})
				require.ErrorIs(t, err, ruletable.ErrPolicyVersionMismatch{})
			})
		}
	})
}
