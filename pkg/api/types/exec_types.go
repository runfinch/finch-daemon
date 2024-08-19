// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import "io"

// ExecConfig is from https://github.com/moby/moby/blob/454b6a7cf5187d1153159e5fe07f4b25c7a95e7f/api/types/configs.go#L30-L45
type ExecConfig struct {
	User         string   // User that will run the command
	Privileged   bool     // Is the container in privileged mode
	Tty          bool     // Attach standard streams to a tty.
	ConsoleSize  *[2]uint `json:",omitempty"` // Initial console size [height, width]
	AttachStdin  bool     // Attach the standard input, makes possible user interaction
	AttachStderr bool     // Attach the standard error
	AttachStdout bool     // Attach the standard output
	Detach       bool     // Execute in detach mode
	DetachKeys   string   // Escape keys for detach
	Env          []string // Environment variables
	WorkingDir   string   // Working directory
	Cmd          []string // Execution commands and args
}

type ExecCreateResponse struct {
	Id string
}

// ExecStartCheck is from https://github.com/moby/moby/blob/454b6a7cf5187d1153159e5fe07f4b25c7a95e7f/api/types/types.go#L231-L240
type ExecStartCheck struct {
	// ExecStart will first check if it's detached
	Detach bool
	// Check if there's a tty
	Tty bool
	// Terminal size [height, width], unused if Tty == false
	ConsoleSize *[2]uint `json:",omitempty"`
}

type ExecStartOptions struct {
	*ExecStartCheck
	ConID           string
	ExecID          string
	Stdin           io.ReadCloser
	Stdout          io.Writer
	Stderr          io.Writer
	SuccessResponse func()
}

type ExecResizeOptions struct {
	ConID  string
	ExecID string
	Height int
	Width  int
}

// ExecInspect holds information about a running process started
// with docker exec.
// from https://github.com/moby/moby/blob/454b6a7cf5187d1153159e5fe07f4b25c7a95e7f/api/types/backend/backend.go#L77-L91
type ExecInspect struct {
	ID            string
	Running       bool
	ExitCode      *int
	ProcessConfig *ExecProcessConfig
	OpenStdin     bool
	OpenStderr    bool
	OpenStdout    bool
	CanRemove     bool
	ContainerID   string
	DetachKeys    []byte
	Pid           int
}

// ExecProcessConfig holds information about the exec process
// running on the host.
// from https://github.com/moby/moby/blob/454b6a7cf5187d1153159e5fe07f4b25c7a95e7f/api/types/backend/backend.go#L93-L101
type ExecProcessConfig struct {
	Tty        bool     `json:"tty"`
	Entrypoint string   `json:"entrypoint"`
	Arguments  []string `json:"arguments"`
	Privileged *bool    `json:"privileged,omitempty"`
	User       string   `json:"user,omitempty"`
}
