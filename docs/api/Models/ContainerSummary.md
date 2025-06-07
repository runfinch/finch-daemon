# ContainerSummary
## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
| **Id** | **String** | Container ID | [optional] [default to null] |
| **Names** | **List** | Container names | [optional] [default to null] |
| **Image** | **String** | Container image | [optional] [default to null] |
| **ImageID** | **String** | Container image ID | [optional] [default to null] |
| **Command** | **String** | Command executed in the container | [optional] [default to null] |
| **Created** | **Long** | When the container was created | [optional] [default to null] |
| **State** | **String** | Container state (created, running, paused, stopped, exited, pausing, unknown) | [optional] [default to null] |
| **Status** | **String** | Container status | [optional] [default to null] |
| **Ports** | [**List**](ContainerSummary_Ports_inner.md) | Published ports | [optional] [default to null] |
| **Labels** | **Map** | Container labels | [optional] [default to null] |
| **NetworkSettings** | [**ContainerSummary_NetworkSettings**](ContainerSummary_NetworkSettings.md) |  | [optional] [default to null] |
| **Mounts** | [**List**](ContainerSummary_Mounts_inner.md) | Container mounts | [optional] [default to null] |

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

