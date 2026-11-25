// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package svc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authzenv1 "github.com/cerbos/cerbos/api/genpb/authzen/authorization/v1"
	enginev1 "github.com/cerbos/cerbos/api/genpb/cerbos/engine/v1"
	requestv1 "github.com/cerbos/cerbos/api/genpb/cerbos/request/v1"
)

func TestValidateRequestActionsAllowsNormalAndWildcards(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateRequestActions([]string{"view", "view:public", "*"}))
}

func TestCheckResourcesRejectsBlankActionBeforeExecution(t *testing.T) {
	t.Parallel()

	svc := &CerbosService{}

	for _, action := range []string{"", " \t\n"} {
		resp, err := svc.CheckResources(context.Background(), &requestv1.CheckResourcesRequest{
			Resources: []*requestv1.CheckResourcesRequest_ResourceEntry{{
				Actions:  []string{action},
				Resource: &enginev1.Resource{Kind: "album", Id: "A10"},
			}},
		})

		require.Nil(t, resp)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	}
}

func TestDeprecatedBatchChecksRejectBlankAction(t *testing.T) {
	t.Parallel()

	svc := &CerbosService{}

	crResp, err := svc.CheckResourceBatch(context.Background(), &requestv1.CheckResourceBatchRequest{
		Resources: []*requestv1.CheckResourceBatchRequest_BatchEntry{{
			Actions: []string{" "},
		}},
	})
	require.Nil(t, crResp)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	setResp, err := svc.CheckResourceSet(context.Background(), &requestv1.CheckResourceSetRequest{
		Actions: []string{" "},
	})
	require.Nil(t, setResp)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestAuthZenBlankActionsAreRejectedAtBothEntries(t *testing.T) {
	t.Parallel()

	svc := &AuthzenAuthorizationService{
		reqLimits: RequestLimits{MaxActionsPerResource: 10, MaxResourcesPerRequest: 10},
	}

	resp, err := svc.AccessEvaluation(context.Background(), &authzenv1.AccessEvaluationRequest{
		Action: &authzenv1.Action{Name: " "},
	})
	require.Nil(t, resp)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	batchResp, err := svc.AccessEvaluationBatch(context.Background(), &authzenv1.AccessEvaluationBatchRequest{
		Evaluations: []*authzenv1.AccessEvaluationBatchRequest_Evaluation{{
			Action: &authzenv1.Action{Name: " "},
		}},
	})
	require.Nil(t, batchResp)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
