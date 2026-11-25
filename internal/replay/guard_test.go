// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package replay

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cerbos/cerbos/internal/storage"
)

func TestConsumeWithinGeneration(t *testing.T) {
	guard := New(time.Minute, 0)

	require.True(t, guard.Consume("req-1"))
	require.False(t, guard.Consume("req-1"), "replay of the same request ID must be rejected")

	require.True(t, guard.Consume("req-2"), "a different request ID must be accepted")
	require.False(t, guard.Consume("req-2"))
}

func TestEmptyRequestIDAlwaysAccepted(t *testing.T) {
	guard := New(time.Minute, 0)

	require.True(t, guard.Consume(""))
	require.True(t, guard.Consume(""), "empty request IDs must not be tracked")
}

func TestPolicyChangeStartsNewGeneration(t *testing.T) {
	guard := New(time.Minute, 0)

	require.True(t, guard.Consume("req-1"))
	require.False(t, guard.Consume("req-1"))

	genBefore := guard.Generation()
	guard.OnStorageEvent(storage.Event{Kind: storage.EventAddOrUpdatePolicy})
	require.Equal(t, genBefore+1, guard.Generation())

	require.True(t, guard.Consume("req-1"), "the request ID may be reused in the new generation")
	require.False(t, guard.Consume("req-1"))

	// Schema-only events must not advance the generation.
	genBefore = guard.Generation()
	guard.OnStorageEvent(storage.Event{Kind: storage.EventAddOrUpdateSchema, SchemaFile: "schema.json"})
	require.Equal(t, genBefore, guard.Generation())
	require.False(t, guard.Consume("req-1"))

	// Reload events start a new generation.
	guard.OnStorageEvent(storage.Event{Kind: storage.EventReload})
	require.Equal(t, genBefore+1, guard.Generation())
	require.True(t, guard.Consume("req-1"))
}

func TestReleaseAllowsRetry(t *testing.T) {
	guard := New(time.Minute, 0)

	require.True(t, guard.Consume("req-1"))
	require.False(t, guard.Consume("req-1"))

	guard.Release("req-1")
	require.True(t, guard.Consume("req-1"), "a released request ID must be consumable again")

	// Releasing an unknown or empty ID must be a safe no-op.
	guard.Release("never-consumed")
	guard.Release("")
}

func TestTTLExpiry(t *testing.T) {
	guard := New(10*time.Millisecond, 0)

	require.True(t, guard.Consume("req-1"))
	require.False(t, guard.Consume("req-1"))

	time.Sleep(20 * time.Millisecond)

	require.True(t, guard.Consume("req-1"), "an expired entry must not block reuse")
}

func TestMaxEntriesEvictsOldest(t *testing.T) {
	guard := New(time.Minute, 2)

	require.True(t, guard.Consume("req-1"))
	require.True(t, guard.Consume("req-2"))

	// Filling the capacity evicts the oldest entry (req-1), not the newest.
	require.True(t, guard.Consume("req-3"))
	require.True(t, guard.Consume("req-1"), "the evicted oldest entry must be consumable again")
	require.False(t, guard.Consume("req-3"), "the newest entry must still be tracked")
}
