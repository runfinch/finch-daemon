// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import "github.com/containerd/nerdctl/pkg/inspecttypes/dockercompat"

// NetworkCreateOption is a utility type for setting
// create network configuration on requests via options.
type NetworkCreateOption func(*NetworkCreateRequest)

// NewCreateNetworkRequest is a utility function for instantiating
// create network requests with both required and optional configuration.
func NewCreateNetworkRequest(name string, opts ...NetworkCreateOption) *NetworkCreateRequest {
	request := &NetworkCreateRequest{Name: name}

	for _, opt := range opts {
		opt(request)
	}

	return request
}

// WithDriver configures the name of the network driver to use.
func WithDriver(driver string) NetworkCreateOption {
	return func(r *NetworkCreateRequest) {
		r.Driver = driver
	}
}

// WithInternal enables/disables internal-only access to the network.
func WithInternal(isInternal bool) NetworkCreateOption {
	return func(r *NetworkCreateRequest) {
		r.Internal = isInternal
	}
}

// WithAttachable enables/disables manual container attachment.
func WithAttachable(isAttachable bool) NetworkCreateOption {
	return func(r *NetworkCreateRequest) {
		r.Attachable = isAttachable
	}
}

// WithIngress enables/disables creating ingress network with routing-mesh in swarm mode.
func WithIngress(isIngress bool) NetworkCreateOption {
	return func(r *NetworkCreateRequest) {
		r.Ingress = isIngress
	}
}

// WithIPAM configures the IP Address Management (IPAM) driver configuration
// and options.
func WithIPAM(ipam IPAM) NetworkCreateOption {
	return func(r *NetworkCreateRequest) {
		r.IPAM = ipam
	}
}

// WithEnableIPv6 enables/disables Internet Protocol version 6 (IPv6) networking.
func WithEnableIPv6(isIPv6 bool) NetworkCreateOption {
	return func(r *NetworkCreateRequest) {
		r.EnableIPv6 = isIPv6
	}
}

// WithOptions configures network driver specific options.
func WithOptions(options map[string]string) NetworkCreateOption {
	return func(r *NetworkCreateRequest) {
		r.Options = options
	}
}

// WithLabels configures user-defined key-value metadata on a network.
func WithLabels(labels map[string]string) NetworkCreateOption {
	return func(r *NetworkCreateRequest) {
		r.Labels = labels
	}
}

// NetworkCreateRequest is a data class for simple JSON marshalling/unmarshalling
// of /networks/create messages into HTTP Post requests.
//
// Reference: https://docs.docker.com/engine/api/v1.43/#tag/Network/operation/NetworkCreate
//
// Example:
//
//	{
//	  "Name": "isolated_nw",
//	  "CheckDuplicate": false,
//	  "Driver": "bridge",
//	  "EnableIPv6": true,
//	  "IPAM": {
//	    "Driver": "default",
//	    "Config": [
//	      {
//	        "Subnet": "172.20.0.0/16",
//	        "IPRange": "172.20.10.0/24",
//	        "Gateway": "172.20.10.11"
//	      },
//	      {
//	        "Subnet": "2001:db8:abcd::/64",
//	        "Gateway": "2001:db8:abcd::1011"
//	      }
//	    ],
//	    "Options": {
//	      "foo": "bar"
//	    }
//	  },
//	  "Internal": true,
//	  "Attachable": false,
//	  "Ingress": false,
//	  "Options": {
//	    "com.docker.network.bridge.default_bridge": "true",
//	    "com.docker.network.bridge.enable_icc": "true",
//	    "com.docker.network.bridge.enable_ip_masquerade": "true",
//	    "com.docker.network.bridge.host_binding_ipv4": "0.0.0.0",
//	    "com.docker.network.bridge.name": "docker0",
//	    "com.docker.network.driver.mtu": "1500"
//	  },
//	  "Labels": {
//	    "com.example.some-label": "some-value",
//	    "com.example.some-other-label": "some-other-value"
//	  }
//	}
type NetworkCreateRequest struct {
	// Name is the network's name.
	Name string `json:"Name"`

	// CheckDuplicate specifies to check for networks with duplicate names.
	//
	// Deprecated: The backend (nerdctl) will always check for collisions.
	CheckDuplicate bool `json:"CheckDuplicate"`

	// Driver is the name of the network driver plugin to use.
	Driver string `json:"Driver" default:"bridge"`

	// Internal specifies to restrict external access to the network.
	//
	// Internal networks are not currently supported.
	Internal bool `json:"Internal"`

	// Attachable specifies if a globally scoped network is manually attachable
	// by regular containers from workers in swarm mode.
	//
	// Attachable networks are not currently supported.
	Attachable bool `json:"Attachable"`

	// Ingress specifies if the network should be an ingress network and provide
	// the routing-mesh in swarm mode.
	//
	// Ingress networks are not currently supported.
	Ingress bool `json:"Ingress"`

	// IPAM specifies customer IP Address Management (IPAM) configuration.
	IPAM IPAM `json:"IPAM"`

	// EnableIPv6 specifies to enable IPv6 on the network.
	EnableIPv6 bool `json:"EnableIPv6"`

	// Options specifies network specific options to be used by the drivers.
	Options map[string]string `json:"Options"`

	// Labels are user-defined key-value network metadata
	Labels map[string]string `json:"Labels"`
}

// IPAM is a data class for simple JSON marshalling/unmarshalling
// of IP Address Management (IPAM) network configuration.
//
// Reference: https://github.com/moby/libnetwork/blob/2267b2527259eff27aa330b35de964afbbb4392e/docs/ipam.md
type IPAM struct {
	// Driver is the name of the IPAM driver to use.
	Driver string `json:"Driver" default:"default"`

	// Config is a list of IPAM configuration options.
	Config []map[string]string `json:"Config"`

	// Options are driver-specific options as a key-value mapping.
	Options map[string]string `json:"Options"`
}

// NetworkCreateResponse is a data class for simple JSON marshalling/unmarshalling
// of /networks/create messages into HTTP Post responses.
//
// Reference: https://docs.docker.com/engine/api/v1.43/#tag/Network/operation/NetworkCreate
//
// Example:
//
//	{
//	  "Id": "22be93d5babb089c5aab8dbc369042fad48ff791584ca2da2100db837a1c7c30",
//	  "Warning": ""
//	}
type NetworkCreateResponse struct {
	// ID is the unique identification document for the network that was created.
	ID string `json:"Id"`

	// Warning is used to communicate any issues which occurred during network configuration.
	Warning string `json:"Warning,omitempty"`
}

// NetworkInspectResponse models a single network object in response to /networks/{id}.
type NetworkInspectResponse struct {
	Name string `json:"Name"`
	ID   string `json:"Id"`
	// Created    string         `json:"Created"`
	// Scope      string         `json:"Scope"`
	// Driver     string         `json:"Driver"`
	// EnableIPv6 bool           `json:"EnableIPv6"`
	// Internal   bool           `json:"Internal"`
	// Attachable bool           `json:"Attachable"`
	// Ingress    bool           `json:"Ingress"`
	IPAM dockercompat.IPAM `json:"IPAM,omitempty"`
	// Containers ContainersType `json:"Containers"`
	// Options    OptionsType    `json:"Options"`
	Labels map[string]string `json:"Labels,omitempty"`
}
