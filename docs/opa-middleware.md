# Applying OPA authz policies

This guide provides instructions for setting up [OPA](https://github.com/open-policy-agent/opa) authz policies with the finch-daemon. Authz policies allow users to allowlist or deny certain resources based on policy rules.

## What Is OPA Authz implementation
Open Policy Agent (OPA) is an open-source, general-purpose policy engine that enables unified, context-aware policy enforcement across the entire stack. OPA provides a high-level declarative language, Rego, for specifying policy as code and simple APIs to offload policy decision-making from your software.

In the current implementation, users can use OPA Rego policies to filter API requests at the Daemon level. It's important to note that the current implementation only supports allowlisting of requests. This means you can specify which requests should be allowed, and all others will be denied by default.

## Setting up a policy 

Use the [sample rego](../sample.rego) policy template to build your policy rules. 

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

Once you are ready with your policy document, use the `--enable-opa` flag to tell the finch-daemon to enable the OPA middleware. The daemon will then look for the policy document provided by the `--rego-file` flag.

Note: The `--rego-file` flag is required when `--enable-opa` is set.

Example: 
`sudo bin/finch-daemon --debug --socket-owner $UID --socket-addr /run/finch-test.sock --pidfile /run/finch-test.pid --enable-opa --rego-file /<path-to>/finch-daemon/sample.rego &`