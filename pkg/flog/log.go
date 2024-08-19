// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package flog contains logging-related APIs.
package flog

// Logger should be used to write any logs. No concrete implementations should be directly used.
//
//go:generate mockgen -destination=../mocks/mocks_logger/logger.go -package=mocks_logger -mock_names Logger=Logger . Logger
type Logger interface {
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	SetLevel(level Level)
}

// Level denotes a log level. Check the constants below for more information.
type Level int

const (
	Debug Level = iota
	Panic
)
