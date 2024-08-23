// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package events

import "github.com/containerd/typeurl/v2"

const (
	CompatibleTopicPrefix = "dockercompat"
)

// From https://github.com/moby/moby/blob/v24.0.4/api/types/events/events.go#L31-L47
type Event struct {
	Status   string     `json:"status,omitempty"` // Deprecated: use Action instead.
	ID       string     `json:"id,omitempty"`     // Deprecated: use Actor.ID instead.
	Type     string     `json:"Type"`
	Action   string     `json:"Action"`
	Actor    EventActor `json:"Actor"`
	Scope    string     `json:"scope"`
	Time     int64      `json:"time"`
	TimeNano int64      `json:"timeNano"`
}

type EventActor struct {
	Id         string            `json:"ID"`
	Attributes map[string]string `json:"Attributes"`
}

func init() {
	typeurl.Register(&Event{}, "dockercompat.event")
	typeurl.Register(&EventActor{}, "dockercompat.event.actor")
}
