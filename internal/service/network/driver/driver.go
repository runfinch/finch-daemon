// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/netutil"
	"github.com/runfinch/finch-daemon/api/types"
)

//go:generate mockgen --destination=../../../../mocks/mocks_network/driver.go -package=mocks_network -mock_names DriverHandler=DriverHandler . DriverHandler
type DriverHandler interface {
	HandleCreateOptions(request types.NetworkCreateRequest, options ncTypes.NetworkCreateOptions) (ncTypes.NetworkCreateOptions, error)
	HandlePostCreate(net *netutil.NetworkConfig) (string, error)
	HandleRemove(net *netutil.NetworkConfig) error
}
