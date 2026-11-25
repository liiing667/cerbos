// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package svc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	enginev1 "github.com/cerbos/cerbos/api/genpb/cerbos/engine/v1"
	requestv1 "github.com/cerbos/cerbos/api/genpb/cerbos/request/v1"
)

func TestValidateRequestActions(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateRequestActions([]string{"view", "edit", "view:*", "*"}))

	err := validateRequestActions([]string{"view", " view"})
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Contains(t, status.Convert(err).Message(), `" view"`)
	require.Contains(t, status.Convert(err).Message(), `"view"`)
}

func TestCheckResourcesRejectsAmbiguousActionsBeforeEvaluation(t *testing.T) {
	t.Parallel()

	svc := &CerbosService{
		reqLimits: RequestLimits{
			MaxActionsPerResource:  10,
			MaxResourcesPerRequest: 10,
		},
	}

	resp, err := svc.CheckResources(t.Context(), &requestv1.CheckResourcesRequest{
		RequestId: "ambiguous-actions",
		Principal: &enginev1.Principal{Id: "alice", Roles: []string{"user"}},
		Resources: []*requestv1.CheckResourcesRequest_ResourceEntry{{
			Actions:  []string{"view", " view"},
			Resource: &enginev1.Resource{Kind: "album", Id: "A1"},
		}},
	})

	require.Nil(t, resp)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
