package finch.authz

import future.keywords.if
import rego.v1

default allow = false

allow if {
    not is_container_create
    not is_networs_api
    not is_swarm_api
    not is_plugins
}

is_container_create if {
    input.Method == "POST"
    glob.match("/**/containers/create", ["/"], input.Path)
}

is_networs_api if {
    input.Method == "GET"
    glob.match("/**/networks", ["/"], input.Path)
}

is_swarm_api if {
    input.Method == "GET"
    glob.match("/**/swarm", ["/"], input.Path)
}

is_plugins if {
    input.Method == "GET"
    glob.match("/**/plugins", ["/"], input.Path)
}

is_forbidden_container if {
    input.Method == "GET"
    glob.match("/**/container/1f576a797a486438548377124f6cb7770a5cb7c8ff6a11c069cb4128d3f59462/json", ["/"], input.Path)
}