// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/runfinch/finch-daemon/pkg/api/types"
	"github.com/runfinch/finch-daemon/pkg/flog"
	"github.com/runfinch/finch-daemon/pkg/version"
)

func (s *service) GetVersion(ctx context.Context) (*types.VersionInfo, error) {
	vInfo := types.VersionInfo{
		Platform:      struct{ Name string }{Name: GetPlatformName()},
		Version:       version.Version,
		ApiVersion:    version.DefaultApiVersion,
		MinAPIVersion: version.MinimumApiVersion,
		GitCommit:     version.GitCommit,
		Os:            getCmdOutput(s.logger, "uname"),
		Arch:          getCmdOutput(s.logger, "uname", "-m"),
		KernelVersion: getCmdOutput(s.logger, "uname", "-r"),
		Experimental:  false,
	}
	sv, err := s.ncSystemSvc.GetServerVersion(ctx)
	if err != nil {
		s.logger.Warnf("unable to retrieve server component versions: %v", err)

		return nil, err
	}
	for _, c := range sv.Components {
		vInfo.Components = append(vInfo.Components, types.ComponentVersion{
			Name:    c.Name,
			Version: c.Version,
			Details: c.Details,
		})
	}
	return &vInfo, nil
}

func getCmdOutput(logger flog.Logger, name string, arg ...string) string {

	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		logger.Warnf("unable to execute command:%s, error: %v", name, err)
		return ""
	}
	return strings.Trim(string(out), "\n")
}

func GetPlatformName() string {
	return fmt.Sprintf("Finch Daemon - %v", version.Version)
}
