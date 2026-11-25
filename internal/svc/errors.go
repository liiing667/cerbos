// Copyright 2021-2026 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package svc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cerbos/cerbos/internal/evaluator"
)

// unknownScopeStatus converts an unknown scope error into a client error that preserves the scope and policy version.
func unknownScopeStatus(err error) error {
	var unknownScopeErr *evaluator.UnknownScopeError
	if errors.As(err, &unknownScopeErr) {
		return status.Error(codes.InvalidArgument, unknownScopeErr.Error())
	}

	return nil
}
