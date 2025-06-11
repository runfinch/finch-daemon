# OPA Authorization Middleware (Experimental)

> ⚠️ **Experimental Feature**: The OPA authorization middleware is being introduced as an experimental feature.

This guide provides instructions for setting up [OPA](https://github.com/open-policy-agent/opa) authorization policies with the finch-daemon. These policies allow users to allowlist or deny certain resources based on policy rules.

## Experimental Status

This feature is being released as experimental because:
- Integration patterns and best practices are still being established
- Performance characteristics are being evaluated

As an experimental feature:
- Breaking changes may occur in any release
- Long-term backward compatibility is not guaranteed
- Documentation and examples may evolve substantially
- Production use is not recommended at this stage

## What Is OPA Authz implementation
Open Policy Agent (OPA) is an open-source, general-purpose policy engine that enables unified, context-aware policy enforcement across the entire stack. OPA provides a high-level declarative language, Rego, for specifying policy as code and simple APIs to offload policy decision-making from your software.

In the current implementation, users can use OPA Rego policies to filter API requests at the Daemon level. It's important to note that the current implementation only supports allowlisting of requests. This means you can specify which requests should be allowed, and all others will be denied by default.

## Setting up a policy 

Use the [sample rego](../docs/sample-rego-policies/example.rego) policy template to build your policy rules. 

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

Once you are ready with your policy document, use the `--experimental` flag to enable experimental features including OPA middleware. The daemon will then look for the policy document provided by the `--rego-file` flag.

Note: Since OPA middleware is an experimental feature, the `--experimental` flag is required when using `--rego-file`.

The daemon enforces strict permissions (0600 or more restrictive) on the Rego policy file to prevent unauthorized modifications. You can bypass this check using the `--skip-rego-perm-check` flag.

Examples:

Standard secure usage:
```bash
sudo bin/finch-daemon --debug --socket-owner $UID --socket-addr /run/finch-test.sock --pidfile /run/finch-test.pid --experimental --rego-file /path/to/policy.rego
```

With permission check bypassed:
```bash
sudo bin/finch-daemon --debug --socket-owner $UID --socket-addr /run/finch-test.sock --pidfile /run/finch-test.pid --experimental --rego-file /path/to/policy.rego --skip-rego-perm-check
```

Note: If you enable experimental features with `--experimental` but don't provide a `--rego-file`, the daemon will run without OPA policy evaluation.


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


### Critical Security Considerations: Rego Policy File Protection

### Rego File Permissions
By default, the daemon requires the Rego policy file to have permissions no more permissive than 0600 (readable and writable only by the owner). This restriction helps prevent unauthorized modifications to the policy file.

The `--skip-rego-perm-check` flag can be used to bypass this permission check. However, using this flag comes with significant security risks:
- More permissive file permissions could allow unauthorized users to modify the policy
- Changes to the policy file could go unnoticed
- Security controls could be weakened without proper oversight

It is strongly recommended to:
- Avoid using `--skip-rego-perm-check` in production environments
- Always use proper file permissions (0600 or more restrictive)
- Implement additional monitoring if the flag must be used

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
