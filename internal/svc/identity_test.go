// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package svc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	enginev1 "github.com/cerbos/cerbos/api/genpb/cerbos/engine/v1"
	requestv1 "github.com/cerbos/cerbos/api/genpb/cerbos/request/v1"
	"github.com/cerbos/cerbos/internal/engine"
	"github.com/cerbos/cerbos/internal/svc"
)

func TestCheckResourcesRejectsBidiPrincipalID(t *testing.T) {
	eng := engine.NewEphemeral(nil, nil, nil)
	cerbosSvc := svc.NewCerbosService(eng, nil, svc.RequestLimits{
		MaxActionsPerResource:  1,
		MaxResourcesPerRequest: 1,
	})

	_, err := cerbosSvc.CheckResources(context.Background(), &requestv1.CheckResourcesRequest{
		RequestId: "identity-test",
		Principal: &enginev1.Principal{
			Id:    "alice\u202e",
			Roles: []string{"user"},
		},
		Resources: []*requestv1.CheckResourcesRequest_ResourceEntry{{
			Actions: []string{"view"},
			Resource: &enginev1.Resource{
				Kind: "album",
				Id:   "A1",
			},
		}},
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Contains(t, st.Message(), `alice\u202e`)
	require.Contains(t, st.Message(), "E2 80 AE")
}
