// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package svc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	enginev1 "github.com/cerbos/cerbos/api/genpb/cerbos/engine/v1"
	requestv1 "github.com/cerbos/cerbos/api/genpb/cerbos/request/v1"
	"github.com/cerbos/cerbos/internal/auxdata"
	"github.com/cerbos/cerbos/internal/compile"
	"github.com/cerbos/cerbos/internal/engine"
	"github.com/cerbos/cerbos/internal/replay"
	"github.com/cerbos/cerbos/internal/ruletable"
	"github.com/cerbos/cerbos/internal/schema"
	"github.com/cerbos/cerbos/internal/storage"
	"github.com/cerbos/cerbos/internal/storage/disk"
	"github.com/cerbos/cerbos/internal/test"
)

func mkService(t *testing.T, guard *replay.Guard) *CerbosService {
	t.Helper()
	ctx := t.Context()

	store, err := disk.NewStore(ctx, &disk.Conf{Directory: test.PathToDir(t, "store")})
	require.NoError(t, err)

	compiler := compile.NewManager(store)

	ruleTable, err := ruletable.NewRuleTableFromLoader(ctx, compiler)
	require.NoError(t, err)

	schemaMgr := schema.NewFromConf(ctx, store, schema.NewConf(schema.EnforcementNone))

	ruletableMgr, err := ruletable.NewRuleTableManager(ruleTable, compiler, schemaMgr)
	require.NoError(t, err)

	eng := engine.NewEphemeral(nil, ruletableMgr, schemaMgr)
	aux := auxdata.NewWithoutVerification(ctx)

	return NewCerbosService(eng, aux, RequestLimits{MaxActionsPerResource: 50, MaxResourcesPerRequest: 50}, guard)
}

func mkCheckResourcesRequest(requestID string) *requestv1.CheckResourcesRequest {
	return &requestv1.CheckResourcesRequest{
		RequestId: requestID,
		Principal: &enginev1.Principal{
			Id:    "donald_duck",
			Roles: []string{"employee"},
		},
		Resources: []*requestv1.CheckResourcesRequest_ResourceEntry{
			{
				Actions: []string{"view:public"},
				Resource: &enginev1.Resource{
					Kind:          "leave_request",
					PolicyVersion: "20210210",
					Id:            "XX125",
					Attr: map[string]*structpb.Value{
						"owner": structpb.NewStringValue("donald_duck"),
					},
				},
			},
		},
	}
}

func TestCheckResourcesReplayProtection(t *testing.T) {
	t.Parallel()

	guard := replay.New(time.Minute, 16)
	svc := mkService(t, guard)
	ctx := t.Context()

	// The first request with a given request ID is evaluated normally.
	resp, err := svc.CheckResources(ctx, mkCheckResourcesRequest("req-1"))
	require.NoError(t, err)
	require.Equal(t, "req-1", resp.RequestId)

	// Replaying the same request ID within the same policy generation is rejected
	// before the engine runs, so no new decision is produced.
	_, err = svc.CheckResources(ctx, mkCheckResourcesRequest("req-1"))
	require.Error(t, err)
	require.Equal(t, codes.AlreadyExists, status.Code(err))

	// A different request ID is still accepted.
	resp, err = svc.CheckResources(ctx, mkCheckResourcesRequest("req-2"))
	require.NoError(t, err)
	require.Equal(t, "req-2", resp.RequestId)

	// Once the policy generation advances, the old ID can be consumed again.
	guard.OnStorageEvent(storage.Event{Kind: storage.EventAddOrUpdatePolicy})
	resp, err = svc.CheckResources(ctx, mkCheckResourcesRequest("req-1"))
	require.NoError(t, err)
	require.Equal(t, "req-1", resp.RequestId)
}

func TestCheckResourcesWithoutReplayProtection(t *testing.T) {
	t.Parallel()

	// A nil guard disables replay protection, preserving the legacy behavior.
	svc := mkService(t, nil)
	ctx := t.Context()

	resp1, err := svc.CheckResources(ctx, mkCheckResourcesRequest("req-1"))
	require.NoError(t, err)
	resp2, err := svc.CheckResources(ctx, mkCheckResourcesRequest("req-1"))
	require.NoError(t, err)
	require.Equal(t, resp1.Results[0].Actions, resp2.Results[0].Actions)
}

func TestCheckResourcesEmptyRequestIDNotTracked(t *testing.T) {
	t.Parallel()

	svc := mkService(t, replay.New(time.Minute, 16))
	ctx := t.Context()

	_, err := svc.CheckResources(ctx, mkCheckResourcesRequest(""))
	require.NoError(t, err)
	_, err = svc.CheckResources(ctx, mkCheckResourcesRequest(""))
	require.NoError(t, err, "empty request IDs must not be subject to replay protection")
}
