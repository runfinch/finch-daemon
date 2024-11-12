// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"os"

	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/netutil"
	"github.com/containernetworking/cni/libcni"
	"github.com/containernetworking/cni/pkg/invoke"
	"github.com/containernetworking/cni/pkg/version"
)

type NerdctlWrapper struct {
	clientWrapper *ContainerdClientWrapper
	globalOptions *types.GlobalCommandOptions
	nerdctlExe    string
	netClient     *netutil.CNIEnv
	CNI           *libcni.CNIConfig
}

func NewNerdctlWrapper(clientWrapper *ContainerdClientWrapper, options *types.GlobalCommandOptions) *NerdctlWrapper {
	return &NerdctlWrapper{
		clientWrapper: clientWrapper,
		globalOptions: options,
		netClient: &netutil.CNIEnv{
			Path:        options.CNIPath,
			NetconfPath: options.CNINetConfPath,
		},
		CNI: libcni.NewCNIConfig(
			[]string{
				options.CNIPath,
			},
			&invoke.DefaultExec{
				RawExec:       &invoke.RawExec{Stderr: os.Stderr},
				PluginDecoder: version.PluginDecoder{},
			}),
	}
}
