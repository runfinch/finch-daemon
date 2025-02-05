package finch.authz

import future.keywords.if
import rego.v1

default allow = false

allow if {
    not is_container_create
}

is_container_create if {
    input.Method == "POST"
    input.Path == "/v1.43/containers/create"
}