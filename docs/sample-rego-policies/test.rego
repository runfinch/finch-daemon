package finch.authz

import future.keywords.if
import rego.v1

default allow = false

allow if {
    not is_container_create
    not is_malformed_api
    not is_containers
}

# Helper rule to ensure path starts with API version
is_api_path if {
    startswith(input.Path, "/v1.43/")
}

# Helper rule to check if path contains containers endpoint regardless of prefix
contains_containers if {
    contains(input.Path, "/containers/")
}

is_container_create if {
    input.Method == "POST"
    contains_containers
    endswith(input.Path, "/create")
}

is_malformed_api if {
    input.Method == "GET"
    contains(input.Path, "/plugins")
}

is_containers if {
    input.Method == "GET"
    contains_containers
    endswith(input.Path, "/json")
}