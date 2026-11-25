// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package index_test

import (
	"testing"

	policyv1 "github.com/cerbos/cerbos/api/genpb/cerbos/policy/v1"
	runtimev1 "github.com/cerbos/cerbos/api/genpb/cerbos/runtime/v1"
	"github.com/cerbos/cerbos/internal/ruletable/index"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNormalizeRoles(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"admin", "user"}, index.NormalizeRoles([]string{"admin", "", "admin", "user", " \t\n"}))
	require.Equal(t, []string{" admin ", "admin"}, index.NormalizeRoles([]string{" admin ", "admin"}))
	require.Empty(t, index.NormalizeRoles([]string{"", " "}))
	require.Equal(t, []string{"admin", "user"}, index.NormalizeRoles([]string{"admin", "user"}))
}

func TestBlankRoleIsNotIndexed(t *testing.T) {
	t.Parallel()

	for _, role := range []string{"", "   ", "\t\n"} {
		impl := index.New()
		require.NoError(t, impl.IndexRule(&runtimev1.RuleTable_RuleRow{
			OriginFqn:  "cerbos.role.admin.vdefault",
			PolicyKind: policyv1.Kind_KIND_RESOURCE,
			Role:       role,
			Resource:   "document",
			Version:    "default",
			Scope:      "",
			ActionSet: &runtimev1.RuleTable_RuleRow_Action{
				Action: "view",
			},
		}))

		require.Empty(t, impl.GetRoles(), "role %q must not create an index identity", role)
	}
}

func TestQueryDeduplicatesRolesWhenSynthesizingRolePolicyDenies(t *testing.T) {
	t.Parallel()

	impl := index.New()
	require.NoError(t, impl.IndexRule(&runtimev1.RuleTable_RuleRow{
		OriginFqn:      "cerbos.role.admin.vdefault",
		PolicyKind:     policyv1.Kind_KIND_RESOURCE,
		FromRolePolicy: true,
		Role:           "admin",
		Resource:       "document",
		Version:        "default",
		Scope:          "",
		ActionSet: &runtimev1.RuleTable_RuleRow_AllowActions_{
			AllowActions: &runtimev1.RuleTable_RuleRow_AllowActions{
				Actions: map[string]*emptypb.Empty{"view": {}},
			},
		},
	}))

	bindings := impl.Query(
		"default",
		index.NewResourceKind("document"),
		"",
		"edit",
		[]string{"admin", "", "admin"},
		policyv1.Kind_KIND_RESOURCE,
		"",
		nil,
	)

	require.Len(t, bindings, 1)
	require.Equal(t, "admin", index.HandleStr(bindings[0].Role))
	require.True(t, bindings[0].NoMatchForScopePermissions)
}
