# ContainerInspectResponse
## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
| **Id** | **String** | Container ID | [optional] [default to null] |
| **Created** | **Date** | When the container was created | [optional] [default to null] |
| **Path** | **String** | Command path | [optional] [default to null] |
| **Args** | **List** | Command arguments | [optional] [default to null] |
| **State** | [**ContainerInspectResponse_State**](ContainerInspectResponse_State.md) |  | [optional] [default to null] |
| **Image** | **String** | Container image | [optional] [default to null] |
| **ResolvConfPath** | **String** | Path to resolv.conf | [optional] [default to null] |
| **HostnamePath** | **String** | Path to hostname | [optional] [default to null] |
| **HostsPath** | **String** | Path to hosts | [optional] [default to null] |
| **LogPath** | **String** | Path to log file | [optional] [default to null] |
| **Name** | **String** | Container name | [optional] [default to null] |
| **RestartCount** | **Integer** | Number of times the container has been restarted | [optional] [default to null] |
| **Driver** | **String** | Storage driver | [optional] [default to null] |
| **Platform** | **String** | Platform | [optional] [default to null] |
| **MountLabel** | **String** | Mount label | [optional] [default to null] |
| **ProcessLabel** | **String** | Process label | [optional] [default to null] |
| **AppArmorProfile** | **String** | AppArmor profile | [optional] [default to null] |
| **ExecIDs** | **List** | Exec IDs | [optional] [default to null] |
| **HostConfig** | [**HostConfig**](HostConfig.md) |  | [optional] [default to null] |
| **GraphDriver** | [**ImageInspect_GraphDriver**](ImageInspect_GraphDriver.md) |  | [optional] [default to null] |
| **SizeRw** | **Long** | Size of writable layer | [optional] [default to null] |
| **SizeRootFs** | **Long** | Size of root filesystem | [optional] [default to null] |
| **Mounts** | [**List**](ContainerSummary_Mounts_inner.md) | Container mounts | [optional] [default to null] |
| **Config** | [**ContainerConfig**](ContainerConfig.md) |  | [optional] [default to null] |
| **NetworkSettings** | [**ContainerInspectResponse_NetworkSettings**](ContainerInspectResponse_NetworkSettings.md) |  | [optional] [default to null] |

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

