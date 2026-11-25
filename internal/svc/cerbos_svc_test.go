// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package svc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	requestv1 "github.com/cerbos/cerbos/api/genpb/cerbos/request/v1"
)

func TestCheckResourcesRequiresRequestID(t *testing.T) {
	svc := &CerbosService{}

	testCases := []struct {
		name string
		req  *requestv1.CheckResourcesRequest
	}{
		{name: "nil request", req: nil},
		{name: "empty request ID", req: &requestv1.CheckResourcesRequest{}},
		{name: "whitespace request ID", req: &requestv1.CheckResourcesRequest{RequestId: " \t"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CheckResources(context.Background(), tc.req)
			require.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}
