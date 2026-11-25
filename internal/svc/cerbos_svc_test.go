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

func TestCheckDuplicateResources(t *testing.T) {
	t.Run("distinct resource identities", func(t *testing.T) {
		resources := []*requestv1.CheckResourcesRequest_ResourceEntry{
			{Resource: &enginev1.Resource{Kind: "leave_request", Id: "XX125"}},
			{Resource: &enginev1.Resource{Kind: "leave_request", Id: "YY125"}},
			{Resource: &enginev1.Resource{Kind: "salary_record", Id: "XX125"}},
		}

		require.NoError(t, checkDuplicateResources(resources))
	})

	t.Run("duplicate resource identity", func(t *testing.T) {
		resources := []*requestv1.CheckResourcesRequest_ResourceEntry{
			{Resource: &enginev1.Resource{Kind: "leave_request", Id: "XX125"}},
			{Resource: &enginev1.Resource{Kind: "leave_request", Id: "YY125"}},
			{Resource: &enginev1.Resource{Kind: "leave_request", Id: "XX125"}},
		}

		err := checkDuplicateResources(resources)
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.Contains(t, status.Convert(err).Message(), "duplicate resource")
	})
}

func TestCerbosServiceRejectsDuplicateResourcesBeforeEvaluation(t *testing.T) {
	svc := &CerbosService{
		reqLimits: RequestLimits{
			MaxActionsPerResource:  50,
			MaxResourcesPerRequest: 50,
		},
	}

	t.Run("check resources", func(t *testing.T) {
		req := &requestv1.CheckResourcesRequest{
			Principal: &enginev1.Principal{Id: "john", Roles: []string{"employee"}},
			Resources: []*requestv1.CheckResourcesRequest_ResourceEntry{
				{
					Actions:  []string{"view"},
					Resource: &enginev1.Resource{Kind: "leave_request", Id: "XX125"},
				},
				{
					Actions:  []string{"approve"},
					Resource: &enginev1.Resource{Kind: "leave_request", Id: "XX125"},
				},
			},
		}

		resp, err := svc.CheckResources(t.Context(), req)
		require.Nil(t, resp)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("check resource batch", func(t *testing.T) {
		req := &requestv1.CheckResourceBatchRequest{
			Principal: &enginev1.Principal{Id: "john", Roles: []string{"employee"}},
			Resources: []*requestv1.CheckResourceBatchRequest_BatchEntry{
				{
					Actions:  []string{"view"},
					Resource: &enginev1.Resource{Kind: "leave_request", Id: "XX125"},
				},
				{
					Actions:  []string{"approve"},
					Resource: &enginev1.Resource{Kind: "leave_request", Id: "XX125"},
				},
			},
		}

		resp, err := svc.CheckResourceBatch(t.Context(), req)
		require.Nil(t, resp)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}
