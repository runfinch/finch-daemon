# ImageInspect
## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
| **Id** | **String** | Image ID | [optional] [default to null] |
| **RepoTags** | **List** | Repository tags | [optional] [default to null] |
| **RepoDigests** | **List** | Repository digests | [optional] [default to null] |
| **Parent** | **String** | Parent image ID | [optional] [default to null] |
| **Comment** | **String** | Comment | [optional] [default to null] |
| **Created** | **Date** | When the image was created | [optional] [default to null] |
| **Container** | **String** | Container used to create the image | [optional] [default to null] |
| **ContainerConfig** | [**ContainerConfig**](ContainerConfig.md) |  | [optional] [default to null] |
| **DockerVersion** | **String** | Docker version used to build the image | [optional] [default to null] |
| **Author** | **String** | Author of the image | [optional] [default to null] |
| **Config** | [**ContainerConfig**](ContainerConfig.md) |  | [optional] [default to null] |
| **Architecture** | **String** | Hardware architecture | [optional] [default to null] |
| **Os** | **String** | Operating system | [optional] [default to null] |
| **OsVersion** | **String** | Operating system version | [optional] [default to null] |
| **Size** | **Long** | Size of the image in bytes | [optional] [default to null] |
| **VirtualSize** | **Long** | Virtual size of the image in bytes | [optional] [default to null] |
| **GraphDriver** | [**ImageInspect_GraphDriver**](ImageInspect_GraphDriver.md) |  | [optional] [default to null] |
| **RootFS** | [**ImageInspect_RootFS**](ImageInspect_RootFS.md) |  | [optional] [default to null] |
| **Metadata** | [**ImageInspect_Metadata**](ImageInspect_Metadata.md) |  | [optional] [default to null] |

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

