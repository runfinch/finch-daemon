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

is_forbidden_container if {
    input.Method == "GET"
    glob.match("/**/containers/1f576a797a486438548377124f6cb7770a5cb7c8ff6a11c069cb4128d3f59462/top", ["/"], input.Path)
}

is_missing_container if {
    input.Method == "GET"
    glob.match("/**/containers/1f576a797a486438548377124f6cb7770a5cb7c8ff6a11c069cb4128d3f59462/json", ["/"], input.Path)
}