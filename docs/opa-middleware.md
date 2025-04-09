# Applying OPA authz policies

This guide provides instructions for setting up [OPA](https://github.com/open-policy-agent/opa) authz policies with the finch-daemon. Authz policies allow users to allowlist or deny certain resources based on policy rules.

## What Is OPA Authz implementation
Open Policy Agent (OPA) is an open-source, general-purpose policy engine that enables unified, context-aware policy enforcement across the entire stack. OPA provides a high-level declarative language, Rego, for specifying policy as code and simple APIs to offload policy decision-making from your software.

In the current implementation, users can use OPA Rego policies to filter API requests at the Daemon level. It's important to note that the current implementation only supports allowlisting of requests. This means you can specify which requests should be allowed, and all others will be denied by default.

## Setting up a policy 

Use the [sample rego](../docs/sample-rego-policies/default.rego) policy template to build your policy rules. 

The package name must be `finch.authz`, the daemon middleware will look for the result of the `allow` key on each API call to determine wether to allow/deny the request. 
An approved request will go through without any events, a rejected request will fail with status code 403

Example: 

The following policy blocks all API requests made to the daemon. 
```
package finch.authz

default allow = false

```
`allow` can be modified based on the business requirements for example we can prevent users from creating new containers by preventing them from accessing the create API

```
allow if {
    not (input.Method == "POST" and input.Path == "/v1.43/containers/create")
}
```
Use the [Rego playground](https://play.openpolicyagent.org/) to fine tune your rego policies

## Enable OPA Middleware

Once you are ready with your policy document, use the `--enable-middleware` flag to tell the finch-daemon to enable the OPA middleware. The daemon will then look for the policy document provided by the `--rego-file` flag.

Note: The `--rego-file` flag is required when `--enable-middleware` is set.

Example: 
`sudo bin/finch-daemon --debug --socket-owner $UID --socket-addr /run/finch-test.sock --pidfile /run/finch-test.pid --enable-middleware --rego-file /<path-to>/finch-daemon/docs/sample-rego-policies/default.rego &`


# Best practices for secure rego policies

## Comprehensive API Path Protection

When writing Rego policies, use pattern matching for API paths to prevent unauthorized access. Simple string matching can be bypassed by adding prefixes to API paths.

Consider this potentially vulnerable policy that tries to restrict access to a specific container:
```
# INCORRECT: Can be bypassed
allow if {
    not (input.Path == "/v1.43/containers/sensitive-container/json")
}
```
This policy can be bypassed in multiple ways:
1. Using container ID instead of name: `/v1.43/containers/abc123.../json`
2. Adding path prefixes: `/custom/v1.43/containers/sensitive-container/json`

Follow the path matching best practices below to properly secure your resources.

## Path Matching Best Practices

```
package finch.authz

import future.keywords.if
import rego.v1

# Use pattern matching for comprehensive path protection
is_container_api if {
    glob.match("/*/containers/*", [], input.Path)
}

is_container_create if {
    input.Method == "POST"
    glob.match("/*/containers/create", [], input.Path)
}

# Protect against path variations
allow if {
    not is_container_api  # Blocks all container-related paths
    not is_container_create  # Specifically blocks container creation
}
```
Use these [example policies](https://github.com/open-policy-agent/opa-docker-authz/blob/2c7eb5c729fca70a3e5cda6f15c2d9cc121b9481/example.rego) to build your opa policy

Remember that only `Method` and `Path` is the only values that 
gets passed to the opa middleware.


### Common Security Pitfalls

- **Incomplete Path Matching**: Always use pattern matching functions like glob.match() instead of exact string matching to catch path variations.
- **Missing HTTP Methods**: Consider all HTTP methods that could access a resource (GET, POST, PUT, DELETE).
- **Alternative API Endpoints**: Be aware that some operations can be performed through multiple endpoints.

### Monitoring and Alerting
The finch-daemon's inability to start due to policy issues could impact system operations. Implement System Service Monitoring in order to be on top of any such failures.

### Security Recommendations
- Policy Testing
  - Test policies in a non-production environment
  - Use the [rego playground](https://play.openpolicyagent.org/) to test policies
- Logging and Audit
  - Enable comprehensive logging of policy decisions
  - Monitor for unexpected denials


### Critical Security Considerations : Rego Policy File Protection
The Rego policy file is a critical security control. 
Any user with sudo privileges can:

- Modify the policy file to weaken security controls
- Replace the policy with a more permissive version
- Disable policy enforcement entirely

#### Recomended Security Controls

- Access Controls
  - Restrict sudo access to specific commands 
- Monitoring
  - Monitor policy file changes
  - Monitor daemon service status