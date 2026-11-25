// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package svc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cerbos/cerbos/internal/ruletable"
)

func TestMapScopeMismatchError_Regression(t *testing.T) {
	err := mapEngineError(ruletable.ErrScopeMismatch{
		PrincipalID:     "alice",
		ResourceID:      "doc-1",
		PrincipalScope:  "tenant-b",
		ResourceScope:   "tenant-a",
		PrincipalTenant: "tenant-b",
		ResourceTenant:  "tenant-a",
	})

	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Contains(t, status.Convert(err).Message(), "tenant-a")
	require.Contains(t, status.Convert(err).Message(), "tenant-b")
}
