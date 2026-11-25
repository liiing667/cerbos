// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package replay guards against decision requests being replayed.
//
// A request identity (the client-supplied request_id) can only be consumed
// once within a single policy generation. When the active policies change, the
// generation advances and the consumed identities are discarded, so the same
// request_id may be used again against the new policy set. Entries also expire
// after a configurable TTL to bound memory usage within a long-lived
// generation.
package replay

import (
	"sync"
	"time"

	"github.com/cerbos/cerbos/internal/storage"
)

// initialGeneration is the policy generation assumed before any policy
// change event has been observed.
const initialGeneration uint64 = 1

// consumedEntry records when a request identity was consumed so that it can be
// expired once it outlives the configured TTL.
type consumedEntry struct {
	consumedAt time.Time
}

// Guard tracks request identities consumed within each policy generation.
// It is safe for concurrent use.
type Guard struct {
	mu         sync.Mutex
	generation uint64
	consumed   map[string]consumedEntry
	ttl        time.Duration
	maxEntries int
}

// New creates a Guard. ttl bounds the lifetime of a consumed identity within a
// generation and maxEntries bounds the number of identities remembered per
// generation; a non-positive maxEntries disables the size bound.
func New(ttl time.Duration, maxEntries int) *Guard {
	return &Guard{
		generation: initialGeneration,
		consumed:   make(map[string]consumedEntry),
		ttl:        ttl,
		maxEntries: maxEntries,
	}
}

// Consume atomically claims the given request identity for the current policy
// generation. It returns true when the identity had not been consumed in this
// generation and is now claimed, and false when it is a replay. An empty
// identity is never tracked and is always accepted.
func (g *Guard) Consume(requestID string) bool {
	if requestID == "" {
		return true
	}

	now := time.Now()

	g.mu.Lock()
	defer g.mu.Unlock()

	g.evictExpiredLocked(now)

	if _, ok := g.consumed[requestID]; ok {
		return false
	}

	// When the bound is reached, drop the oldest entry rather than rejecting a
	// genuinely new request. Replay protection is best-effort within a
	// generation and the TTL already provides a hard retention boundary.
	if g.maxEntries > 0 && len(g.consumed) >= g.maxEntries {
		g.evictOldestLocked()
	}

	g.consumed[requestID] = consumedEntry{consumedAt: now}
	return true
}

// Release undoes a claim made by Consume. It must be called when a request
// failed before producing a decision, so that the client can safely retry with
// the same request identity. It is a no-op for empty identities.
func (g *Guard) Release(requestID string) {
	if requestID == "" {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.consumed, requestID)
}

// Generation returns the current policy generation. It is mainly useful for
// tests and observability.
func (g *Guard) Generation() uint64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.generation
}

// OnStorageEvent advances the policy generation and forgets every identity
// consumed in previous generations whenever the policy set changes.
func (g *Guard) OnStorageEvent(events ...storage.Event) {
	var policyChanged bool
	for _, event := range events {
		switch event.Kind {
		case storage.EventAddOrUpdatePolicy,
			storage.EventDeleteOrDisablePolicy,
			storage.EventReload:
			policyChanged = true
		}
	}

	if !policyChanged {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.generation++
	g.consumed = make(map[string]consumedEntry)
}

// SubscriberID implements storage.Subscriber.
func (g *Guard) SubscriberID() string {
	return "replay.Guard"
}

func (g *Guard) evictExpiredLocked(now time.Time) {
	if g.ttl <= 0 {
		return
	}

	for id, entry := range g.consumed {
		if now.Sub(entry.consumedAt) >= g.ttl {
			delete(g.consumed, id)
		}
	}
}

func (g *Guard) evictOldestLocked() {
	var (
		oldestID string
		oldestAt time.Time
		found    bool
	)

	for id, entry := range g.consumed {
		if !found || entry.consumedAt.Before(oldestAt) {
			oldestID, oldestAt, found = id, entry.consumedAt, true
		}
	}

	if found {
		delete(g.consumed, oldestID)
	}
}
