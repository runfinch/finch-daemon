# CreateContainerRequest
## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
| **Hostname** | **String** | Container host name | [optional] [default to null] |
| **User** | **String** | Username or UID to run commands inside the container | [optional] [default to null] |
| **AttachStdin** | **Boolean** | Attach the standard input | [optional] [default to false] |
| **ExposedPorts** | **Map** | Ports to expose from the container without publishing to the host | [optional] [default to null] |
| **Tty** | **Boolean** | Attach standard streams to a TTY, including stdin if it is not closed | [optional] [default to false] |
| **Env** | **List** | List of environment variables in the form [\&quot;VAR&#x3D;value\&quot;, ...] | [optional] [default to null] |
| **Cmd** | **List** | Command to run when starting the container | [optional] [default to null] |
| **Image** | **String** | Name of the image as it was passed by the operator | [default to null] |
| **Volumes** | **Map** | List of volumes (mounts) used for the container | [optional] [default to null] |
| **WorkingDir** | **String** | Current directory (PWD) in the command will be launched | [optional] [default to null] |
| **Entrypoint** | **List** | Entrypoint to run when starting the container | [optional] [default to null] |
| **NetworkDisabled** | **Boolean** | Disable networking for the container | [optional] [default to null] |
| **MacAddress** | **String** | Container MAC address (e.g., \&quot;12:34:56:78:9a:bc\&quot;) | [optional] [default to null] |
| **Labels** | **Map** | Key-value map of container metadata | [optional] [default to null] |
| **StopSignal** | **String** | Signal to stop a container (e.g., \&quot;SIGTERM\&quot;) | [optional] [default to null] |
| **StopTimeout** | **Integer** | Timeout (in seconds) to stop a container | [optional] [default to null] |
| **HostConfig** | [**HostConfig**](HostConfig.md) |  | [optional] [default to null] |

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

