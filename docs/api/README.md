# Documentation for Finch Daemon API

<a name="documentation-for-api-endpoints"></a>
## Documentation for API Endpoints

All URIs are relative to *http://localhost/v1.43*

| Class | Method | HTTP request | Description |
|------------ | ------------- | ------------- | -------------|
| *DefaultApi* | [**containerArchive**](Apis/DefaultApi.md#containerarchive) | **GET** /containers/{id}/archive | Get an archive of a filesystem resource in a container |
*DefaultApi* | [**containerArchiveExtract**](Apis/DefaultApi.md#containerarchiveextract) | **PUT** /containers/{id}/archive | Extract an archive of files or folders to a directory in a container |
*DefaultApi* | [**containerArchiveInfo**](Apis/DefaultApi.md#containerarchiveinfo) | **HEAD** /containers/{id}/archive | Get information about files in a container |
*DefaultApi* | [**containerAttach**](Apis/DefaultApi.md#containerattach) | **POST** /containers/{id}/attach | Attach to a container |
*DefaultApi* | [**containerCreate**](Apis/DefaultApi.md#containercreate) | **POST** /containers/create | Create a container |
*DefaultApi* | [**containerDelete**](Apis/DefaultApi.md#containerdelete) | **DELETE** /containers/{id} | Remove a container |
*DefaultApi* | [**containerExec**](Apis/DefaultApi.md#containerexec) | **POST** /containers/{id}/exec | Create an exec instance |
*DefaultApi* | [**containerInspect**](Apis/DefaultApi.md#containerinspect) | **GET** /containers/{id}/json | Inspect a container |
*DefaultApi* | [**containerKill**](Apis/DefaultApi.md#containerkill) | **POST** /containers/{id}/kill | Kill a container |
*DefaultApi* | [**containerList**](Apis/DefaultApi.md#containerlist) | **GET** /containers/json | List containers |
*DefaultApi* | [**containerLogs**](Apis/DefaultApi.md#containerlogs) | **GET** /containers/{id}/logs | Get container logs |
*DefaultApi* | [**containerPause**](Apis/DefaultApi.md#containerpause) | **POST** /containers/{id}/pause | Pause a container |
*DefaultApi* | [**containerRemove**](Apis/DefaultApi.md#containerremove) | **POST** /containers/{id}/remove | Remove a container |
*DefaultApi* | [**containerRename**](Apis/DefaultApi.md#containerrename) | **POST** /containers/{id}/rename | Rename a container |
*DefaultApi* | [**containerRestart**](Apis/DefaultApi.md#containerrestart) | **POST** /containers/{id}/restart | Restart a container |
*DefaultApi* | [**containerStart**](Apis/DefaultApi.md#containerstart) | **POST** /containers/{id}/start | Start a container |
*DefaultApi* | [**containerStats**](Apis/DefaultApi.md#containerstats) | **GET** /containers/{id}/stats | Get container stats based on resource usage |
*DefaultApi* | [**containerStop**](Apis/DefaultApi.md#containerstop) | **POST** /containers/{id}/stop | Stop a container |
*DefaultApi* | [**containerUnpause**](Apis/DefaultApi.md#containerunpause) | **POST** /containers/{id}/unpause | Unpause a container |
*DefaultApi* | [**containerWait**](Apis/DefaultApi.md#containerwait) | **POST** /containers/{id}/wait | Wait for a container |
*DefaultApi* | [**distributionInspect**](Apis/DefaultApi.md#distributioninspect) | **GET** /distribution/{name}/json | Get image distribution information |
*DefaultApi* | [**execInspect**](Apis/DefaultApi.md#execinspect) | **GET** /exec/{id}/json | Inspect an exec instance |
*DefaultApi* | [**execResize**](Apis/DefaultApi.md#execresize) | **POST** /exec/{id}/resize | Resize an exec instance |
*DefaultApi* | [**execStart**](Apis/DefaultApi.md#execstart) | **POST** /exec/{id}/start | Start an exec instance |
*DefaultApi* | [**imageBuild**](Apis/DefaultApi.md#imagebuild) | **POST** /build | Build an image |
*DefaultApi* | [**imageCreate**](Apis/DefaultApi.md#imagecreate) | **POST** /images/create | Create an image |
*DefaultApi* | [**imageDelete**](Apis/DefaultApi.md#imagedelete) | **DELETE** /images/{name} | Remove an image |
*DefaultApi* | [**imageInspect**](Apis/DefaultApi.md#imageinspect) | **GET** /images/{name}/json | Inspect an image |
*DefaultApi* | [**imageList**](Apis/DefaultApi.md#imagelist) | **GET** /images/json | List images |
*DefaultApi* | [**imageLoad**](Apis/DefaultApi.md#imageload) | **POST** /images/load | Load an image |
*DefaultApi* | [**imagePush**](Apis/DefaultApi.md#imagepush) | **POST** /images/{name}/push | Push an image |
*DefaultApi* | [**imageTag**](Apis/DefaultApi.md#imagetag) | **POST** /images/{name}/tag | Tag an image |
*DefaultApi* | [**networkConnect**](Apis/DefaultApi.md#networkconnect) | **POST** /networks/{id}/connect | Connect a container to a network |
*DefaultApi* | [**networkCreate**](Apis/DefaultApi.md#networkcreate) | **POST** /networks/create | Create a network |
*DefaultApi* | [**networkDelete**](Apis/DefaultApi.md#networkdelete) | **DELETE** /networks/{id} | Remove a network |
*DefaultApi* | [**networkInspect**](Apis/DefaultApi.md#networkinspect) | **GET** /networks/{id} | Inspect a network |
*DefaultApi* | [**networkList**](Apis/DefaultApi.md#networklist) | **GET** /networks | List networks |
*DefaultApi* | [**systemAuth**](Apis/DefaultApi.md#systemauth) | **POST** /auth | Check auth configuration |
*DefaultApi* | [**systemEvents**](Apis/DefaultApi.md#systemevents) | **GET** /events | Monitor events |
*DefaultApi* | [**systemInfo**](Apis/DefaultApi.md#systeminfo) | **GET** /info | Get system information |
*DefaultApi* | [**systemPing**](Apis/DefaultApi.md#systemping) | **GET** /_ping | Ping |
*DefaultApi* | [**systemVersion**](Apis/DefaultApi.md#systemversion) | **GET** /version | Get version |
*DefaultApi* | [**volumeCreate**](Apis/DefaultApi.md#volumecreate) | **POST** /volumes/create | Create a volume |
*DefaultApi* | [**volumeDelete**](Apis/DefaultApi.md#volumedelete) | **DELETE** /volumes/{name} | Remove a volume |
*DefaultApi* | [**volumeInspect**](Apis/DefaultApi.md#volumeinspect) | **GET** /volumes/{name} | Inspect a volume |
*DefaultApi* | [**volumeList**](Apis/DefaultApi.md#volumelist) | **GET** /volumes | List volumes |


<a name="documentation-for-models"></a>
## Documentation for Models

 - [AuthConfig](./Models/AuthConfig.md)
 - [AuthResponse](./Models/AuthResponse.md)
 - [BuildResult](./Models/BuildResult.md)
 - [BuildResult_aux](./Models/BuildResult_aux.md)
 - [BuildResult_errorDetail](./Models/BuildResult_errorDetail.md)
 - [BuildResult_progressDetail](./Models/BuildResult_progressDetail.md)
 - [ContainerConfig](./Models/ContainerConfig.md)
 - [ContainerExec_201_response](./Models/ContainerExec_201_response.md)
 - [ContainerInspectResponse](./Models/ContainerInspectResponse.md)
 - [ContainerInspectResponse_NetworkSettings](./Models/ContainerInspectResponse_NetworkSettings.md)
 - [ContainerInspectResponse_NetworkSettings_Networks_value](./Models/ContainerInspectResponse_NetworkSettings_Networks_value.md)
 - [ContainerInspectResponse_NetworkSettings_Networks_value_IPAMConfig](./Models/ContainerInspectResponse_NetworkSettings_Networks_value_IPAMConfig.md)
 - [ContainerInspectResponse_NetworkSettings_Ports_value_inner](./Models/ContainerInspectResponse_NetworkSettings_Ports_value_inner.md)
 - [ContainerInspectResponse_State](./Models/ContainerInspectResponse_State.md)
 - [ContainerSummary](./Models/ContainerSummary.md)
 - [ContainerSummary_Mounts_inner](./Models/ContainerSummary_Mounts_inner.md)
 - [ContainerSummary_NetworkSettings](./Models/ContainerSummary_NetworkSettings.md)
 - [ContainerSummary_NetworkSettings_Networks_value](./Models/ContainerSummary_NetworkSettings_Networks_value.md)
 - [ContainerSummary_Ports_inner](./Models/ContainerSummary_Ports_inner.md)
 - [ContainerWaitResponse](./Models/ContainerWaitResponse.md)
 - [ContainerWaitResponse_Error](./Models/ContainerWaitResponse_Error.md)
 - [CreateContainerRequest](./Models/CreateContainerRequest.md)
 - [CreateContainerResponse](./Models/CreateContainerResponse.md)
 - [DistributionInspect](./Models/DistributionInspect.md)
 - [DistributionInspect_Descriptor](./Models/DistributionInspect_Descriptor.md)
 - [DistributionInspect_Platforms_inner](./Models/DistributionInspect_Platforms_inner.md)
 - [Error](./Models/Error.md)
 - [EventMessage](./Models/EventMessage.md)
 - [ExecConfig](./Models/ExecConfig.md)
 - [ExecInspectResponse](./Models/ExecInspectResponse.md)
 - [ExecInspectResponse_ProcessConfig](./Models/ExecInspectResponse_ProcessConfig.md)
 - [ExecStartConfig](./Models/ExecStartConfig.md)
 - [HostConfig](./Models/HostConfig.md)
 - [HostConfig_BlkioDeviceReadBps_inner](./Models/HostConfig_BlkioDeviceReadBps_inner.md)
 - [HostConfig_BlkioDeviceReadIOps_inner](./Models/HostConfig_BlkioDeviceReadIOps_inner.md)
 - [HostConfig_BlkioWeightDevice_inner](./Models/HostConfig_BlkioWeightDevice_inner.md)
 - [HostConfig_Devices_inner](./Models/HostConfig_Devices_inner.md)
 - [HostConfig_LogConfig](./Models/HostConfig_LogConfig.md)
 - [HostConfig_PortBindings_value_inner](./Models/HostConfig_PortBindings_value_inner.md)
 - [HostConfig_RestartPolicy](./Models/HostConfig_RestartPolicy.md)
 - [ImageDelete_200_response_inner](./Models/ImageDelete_200_response_inner.md)
 - [ImageInspect](./Models/ImageInspect.md)
 - [ImageInspect_GraphDriver](./Models/ImageInspect_GraphDriver.md)
 - [ImageInspect_Metadata](./Models/ImageInspect_Metadata.md)
 - [ImageInspect_RootFS](./Models/ImageInspect_RootFS.md)
 - [ImageLoadResponse](./Models/ImageLoadResponse.md)
 - [ImageSummary](./Models/ImageSummary.md)
 - [Network](./Models/Network.md)
 - [NetworkConnectRequest](./Models/NetworkConnectRequest.md)
 - [NetworkConnectRequest_EndpointConfig](./Models/NetworkConnectRequest_EndpointConfig.md)
 - [NetworkConnectRequest_EndpointConfig_IPAMConfig](./Models/NetworkConnectRequest_EndpointConfig_IPAMConfig.md)
 - [NetworkCreateRequest](./Models/NetworkCreateRequest.md)
 - [NetworkCreateRequest_IPAM](./Models/NetworkCreateRequest_IPAM.md)
 - [NetworkCreateResponse](./Models/NetworkCreateResponse.md)
 - [Network_ConfigFrom](./Models/Network_ConfigFrom.md)
 - [Network_Containers_value](./Models/Network_Containers_value.md)
 - [Network_IPAM](./Models/Network_IPAM.md)
 - [Network_IPAM_Config_inner](./Models/Network_IPAM_Config_inner.md)
 - [SystemEvents_200_response](./Models/SystemEvents_200_response.md)
 - [SystemEvents_200_response_Actor](./Models/SystemEvents_200_response_Actor.md)
 - [SystemInfo](./Models/SystemInfo.md)
 - [SystemInfo_ContainerdCommit](./Models/SystemInfo_ContainerdCommit.md)
 - [SystemInfo_DefaultAddressPools_inner](./Models/SystemInfo_DefaultAddressPools_inner.md)
 - [SystemInfo_GenericResources_inner](./Models/SystemInfo_GenericResources_inner.md)
 - [SystemInfo_GenericResources_inner_DiscreteResourceSpec](./Models/SystemInfo_GenericResources_inner_DiscreteResourceSpec.md)
 - [SystemInfo_GenericResources_inner_NamedResourceSpec](./Models/SystemInfo_GenericResources_inner_NamedResourceSpec.md)
 - [SystemInfo_InitCommit](./Models/SystemInfo_InitCommit.md)
 - [SystemInfo_Plugins](./Models/SystemInfo_Plugins.md)
 - [SystemInfo_RegistryConfig](./Models/SystemInfo_RegistryConfig.md)
 - [SystemInfo_RegistryConfig_IndexConfigs_value](./Models/SystemInfo_RegistryConfig_IndexConfigs_value.md)
 - [SystemInfo_RuncCommit](./Models/SystemInfo_RuncCommit.md)
 - [SystemInfo_Runtimes_value](./Models/SystemInfo_Runtimes_value.md)
 - [SystemInfo_Swarm](./Models/SystemInfo_Swarm.md)
 - [SystemInfo_Swarm_Cluster](./Models/SystemInfo_Swarm_Cluster.md)
 - [SystemInfo_Swarm_Cluster_Spec](./Models/SystemInfo_Swarm_Cluster_Spec.md)
 - [SystemInfo_Swarm_Cluster_Spec_CAConfig](./Models/SystemInfo_Swarm_Cluster_Spec_CAConfig.md)
 - [SystemInfo_Swarm_Cluster_Spec_CAConfig_ExternalCAs_inner](./Models/SystemInfo_Swarm_Cluster_Spec_CAConfig_ExternalCAs_inner.md)
 - [SystemInfo_Swarm_Cluster_Spec_Dispatcher](./Models/SystemInfo_Swarm_Cluster_Spec_Dispatcher.md)
 - [SystemInfo_Swarm_Cluster_Spec_EncryptionConfig](./Models/SystemInfo_Swarm_Cluster_Spec_EncryptionConfig.md)
 - [SystemInfo_Swarm_Cluster_Spec_Orchestration](./Models/SystemInfo_Swarm_Cluster_Spec_Orchestration.md)
 - [SystemInfo_Swarm_Cluster_Spec_Raft](./Models/SystemInfo_Swarm_Cluster_Spec_Raft.md)
 - [SystemInfo_Swarm_Cluster_Spec_TaskDefaults](./Models/SystemInfo_Swarm_Cluster_Spec_TaskDefaults.md)
 - [SystemInfo_Swarm_Cluster_Spec_TaskDefaults_LogDriver](./Models/SystemInfo_Swarm_Cluster_Spec_TaskDefaults_LogDriver.md)
 - [SystemInfo_Swarm_Cluster_Version](./Models/SystemInfo_Swarm_Cluster_Version.md)
 - [SystemInfo_Swarm_RemoteManagers_inner](./Models/SystemInfo_Swarm_RemoteManagers_inner.md)
 - [VersionResponse](./Models/VersionResponse.md)
 - [VersionResponse_Components_inner](./Models/VersionResponse_Components_inner.md)
 - [VersionResponse_Platform](./Models/VersionResponse_Platform.md)
 - [Volume](./Models/Volume.md)
 - [VolumeCreateRequest](./Models/VolumeCreateRequest.md)
 - [VolumeListResponse](./Models/VolumeListResponse.md)
 - [Volume_UsageData](./Models/Volume_UsageData.md)


<a name="documentation-for-authorization"></a>
## Documentation for Authorization

All endpoints do not require authorization.
