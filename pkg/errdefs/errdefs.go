// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package errdefs contains the error definitions (i.e., domain types).
// It is created to prevent adaptor layers (e.g., HTTP handler) from
// directly depending on the errors received by infra layers (e.g., containerd's error).
package errdefs

import (
	"github.com/pkg/errors"
)

type errType int

const (
	unknown errType = iota
	unauthenticated
	notFound
	invalidFormat
	conflict
	notModified
	wrongSemantics
	forbidden
)

type errWithType struct {
	t       errType
	wrapped error
}

var _ error = &errWithType{}

func (e *errWithType) Error() string {
	return e.wrapped.Error()
}

func (e *errWithType) Unwrap() error {
	return e.wrapped
}

func new(t errType, err error) error {
	return &errWithType{
		t:       t,
		wrapped: err,
	}
}

// isType only checks the first errWithType in the error chain.
func isType(t errType, err error) bool {
	var ewt *errWithType
	if errors.As(err, &ewt) {
		return ewt.t == t
	}
	return false
}

func NewUnauthenticated(err error) error {
	return new(unauthenticated, err)
}

func IsUnauthenticated(err error) bool {
	return isType(unauthenticated, err)
}

func NewNotFound(err error) error {
	return new(notFound, err)
}

func IsNotFound(err error) bool {
	return isType(notFound, err)
}

func NewInvalidFormat(err error) error {
	return new(invalidFormat, err)
}

func IsInvalidFormat(err error) bool {
	return isType(invalidFormat, err)
}

func NewConflict(err error) error {
	return new(conflict, err)
}

func IsConflict(err error) bool {
	return isType(conflict, err)
}

func NewNotModified(err error) error {
	return new(notModified, err)
}
func IsNotModified(err error) bool {
	return isType(notModified, err)
}

func NewWrongSemantics(err error) error {
	return new(wrongSemantics, err)
}

func IsWrongSemantics(err error) bool {
	return isType(wrongSemantics, err)
}

func NewForbidden(err error) error {
	return new(forbidden, err)
}

func IsForbiddenError(err error) bool {
	return isType(forbidden, err)
}
