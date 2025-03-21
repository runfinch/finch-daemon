package finch.authz

import future.keywords.if
import rego.v1

default allow = false

allow if {
    not is_container_create
}

is_container_create if {
    input.Method == "POST"
    glob.match("/**/containers/create", ["/"], input.Path)
}

is_swarm_api if {
    input.Method == "GET"
    glob.match("/**/swarm", ["/"], input.Path)
}
