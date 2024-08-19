// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ecc

import (
	"io"
	"os/exec"
)

//go:generate mockgen --destination=../mocks/mocks_ecc/execcmdcreator.go -package=mocks_ecc github.com/runfinch/finch-daemon/pkg/ecc ExecCmdCreator
type ExecCmdCreator interface {
	Command(name string, args ...string) ExecCmd
}

//go:generate mockgen --destination=../mocks/mocks_ecc/execcmd.go -package=mocks_ecc github.com/runfinch/finch-daemon/pkg/ecc ExecCmd
type ExecCmd interface {
	Run() error
	SetDir(path string)
	SetStdout(writer io.Writer)
	SetStderr(writer io.Writer)
	SetStdin(reader io.Reader)
	GetDir() string
}

func NewExecCmdCreator() ExecCmdCreator {
	return &execCmdCreator{}
}

type execCmdCreator struct {
}

func (*execCmdCreator) Command(name string, args ...string) ExecCmd {
	return &execCmd{
		cmd: exec.Command(name, args...),
	}
}

type execCmd struct {
	cmd *exec.Cmd
}

// Run runs the command
func (c *execCmd) Run() error {
	return c.cmd.Run()
}

// SetDir sets the command's working directory
func (c *execCmd) SetDir(path string) {
	c.cmd.Dir = path
}

// SetStdout sets the command's standard output.
func (c *execCmd) SetStdout(writer io.Writer) {
	c.cmd.Stdout = writer
}

// SetStderr sets the command's standard error output.
func (c *execCmd) SetStderr(writer io.Writer) {
	c.cmd.Stderr = writer
}

// SetStdin sets the command's standard input.
func (c *execCmd) SetStdin(reader io.Reader) {
	c.cmd.Stdin = reader
}

// GetDir gets the command's working directory
func (c *execCmd) GetDir() string {
	return c.cmd.Dir
}
