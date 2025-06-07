# ContainerInspectResponse_State
## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
| **Status** | **String** | Container status (created, running, paused, restarting, removing, exited, dead) | [optional] [default to null] |
| **Running** | **Boolean** | Whether the container is running | [optional] [default to null] |
| **Paused** | **Boolean** | Whether the container is paused | [optional] [default to null] |
| **Restarting** | **Boolean** | Whether the container is restarting | [optional] [default to null] |
| **OOMKilled** | **Boolean** | Whether the container was killed because it ran out of memory | [optional] [default to null] |
| **Dead** | **Boolean** | Whether the container is dead | [optional] [default to null] |
| **Pid** | **Integer** | Process ID | [optional] [default to null] |
| **ExitCode** | **Integer** | Exit code | [optional] [default to null] |
| **Error** | **String** | Error message | [optional] [default to null] |
| **StartedAt** | **Date** | When the container was started | [optional] [default to null] |
| **FinishedAt** | **Date** | When the container finished | [optional] [default to null] |

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

