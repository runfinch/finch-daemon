# DefaultApi

All URIs are relative to *http://localhost/v1.43*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**containerArchive**](DefaultApi.md#containerArchive) | **GET** /containers/{id}/archive | Get an archive of a filesystem resource in a container |
| [**containerArchiveExtract**](DefaultApi.md#containerArchiveExtract) | **PUT** /containers/{id}/archive | Extract an archive of files or folders to a directory in a container |
| [**containerArchiveInfo**](DefaultApi.md#containerArchiveInfo) | **HEAD** /containers/{id}/archive | Get information about files in a container |
| [**containerAttach**](DefaultApi.md#containerAttach) | **POST** /containers/{id}/attach | Attach to a container |
| [**containerCreate**](DefaultApi.md#containerCreate) | **POST** /containers/create | Create a container |
| [**containerDelete**](DefaultApi.md#containerDelete) | **DELETE** /containers/{id} | Remove a container |
| [**containerExec**](DefaultApi.md#containerExec) | **POST** /containers/{id}/exec | Create an exec instance |
| [**containerInspect**](DefaultApi.md#containerInspect) | **GET** /containers/{id}/json | Inspect a container |
| [**containerKill**](DefaultApi.md#containerKill) | **POST** /containers/{id}/kill | Kill a container |
| [**containerList**](DefaultApi.md#containerList) | **GET** /containers/json | List containers |
| [**containerLogs**](DefaultApi.md#containerLogs) | **GET** /containers/{id}/logs | Get container logs |
| [**containerPause**](DefaultApi.md#containerPause) | **POST** /containers/{id}/pause | Pause a container |
| [**containerRemove**](DefaultApi.md#containerRemove) | **POST** /containers/{id}/remove | Remove a container |
| [**containerRename**](DefaultApi.md#containerRename) | **POST** /containers/{id}/rename | Rename a container |
| [**containerRestart**](DefaultApi.md#containerRestart) | **POST** /containers/{id}/restart | Restart a container |
| [**containerStart**](DefaultApi.md#containerStart) | **POST** /containers/{id}/start | Start a container |
| [**containerStats**](DefaultApi.md#containerStats) | **GET** /containers/{id}/stats | Get container stats based on resource usage |
| [**containerStop**](DefaultApi.md#containerStop) | **POST** /containers/{id}/stop | Stop a container |
| [**containerUnpause**](DefaultApi.md#containerUnpause) | **POST** /containers/{id}/unpause | Unpause a container |
| [**containerWait**](DefaultApi.md#containerWait) | **POST** /containers/{id}/wait | Wait for a container |
| [**distributionInspect**](DefaultApi.md#distributionInspect) | **GET** /distribution/{name}/json | Get image distribution information |
| [**execInspect**](DefaultApi.md#execInspect) | **GET** /exec/{id}/json | Inspect an exec instance |
| [**execResize**](DefaultApi.md#execResize) | **POST** /exec/{id}/resize | Resize an exec instance |
| [**execStart**](DefaultApi.md#execStart) | **POST** /exec/{id}/start | Start an exec instance |
| [**imageBuild**](DefaultApi.md#imageBuild) | **POST** /build | Build an image |
| [**imageCreate**](DefaultApi.md#imageCreate) | **POST** /images/create | Create an image |
| [**imageDelete**](DefaultApi.md#imageDelete) | **DELETE** /images/{name} | Remove an image |
| [**imageInspect**](DefaultApi.md#imageInspect) | **GET** /images/{name}/json | Inspect an image |
| [**imageList**](DefaultApi.md#imageList) | **GET** /images/json | List images |
| [**imageLoad**](DefaultApi.md#imageLoad) | **POST** /images/load | Load an image |
| [**imagePush**](DefaultApi.md#imagePush) | **POST** /images/{name}/push | Push an image |
| [**imageTag**](DefaultApi.md#imageTag) | **POST** /images/{name}/tag | Tag an image |
| [**networkConnect**](DefaultApi.md#networkConnect) | **POST** /networks/{id}/connect | Connect a container to a network |
| [**networkCreate**](DefaultApi.md#networkCreate) | **POST** /networks/create | Create a network |
| [**networkDelete**](DefaultApi.md#networkDelete) | **DELETE** /networks/{id} | Remove a network |
| [**networkInspect**](DefaultApi.md#networkInspect) | **GET** /networks/{id} | Inspect a network |
| [**networkList**](DefaultApi.md#networkList) | **GET** /networks | List networks |
| [**systemAuth**](DefaultApi.md#systemAuth) | **POST** /auth | Check auth configuration |
| [**systemEvents**](DefaultApi.md#systemEvents) | **GET** /events | Monitor events |
| [**systemInfo**](DefaultApi.md#systemInfo) | **GET** /info | Get system information |
| [**systemPing**](DefaultApi.md#systemPing) | **GET** /_ping | Ping |
| [**systemVersion**](DefaultApi.md#systemVersion) | **GET** /version | Get version |
| [**volumeCreate**](DefaultApi.md#volumeCreate) | **POST** /volumes/create | Create a volume |
| [**volumeDelete**](DefaultApi.md#volumeDelete) | **DELETE** /volumes/{name} | Remove a volume |
| [**volumeInspect**](DefaultApi.md#volumeInspect) | **GET** /volumes/{name} | Inspect a volume |
| [**volumeList**](DefaultApi.md#volumeList) | **GET** /volumes | List volumes |


<a name="containerArchive"></a>
# **containerArchive**
> File containerArchive(id, path)

Get an archive of a filesystem resource in a container

    Get a tar archive of a resource in the filesystem of a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **path** | **String**| Resource path in the container | [default to null] |

### Return type

**File**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream, application/json

<a name="containerArchiveExtract"></a>
# **containerArchiveExtract**
> containerArchiveExtract(id, path, body, noOverwriteDirNonDir)

Extract an archive of files or folders to a directory in a container

    Upload a tar archive to be extracted to a path in the filesystem of a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **path** | **String**| Resource path in the container | [default to null] |
| **body** | **File**| The input stream must be a tar archive compressed with one of the following algorithms gzip, bzip2, xz | |
| **noOverwriteDirNonDir** | **String**| Do not overwrite directory with non-directory and vice versa | [optional] [default to null] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/octet-stream
- **Accept**: application/json

<a name="containerArchiveInfo"></a>
# **containerArchiveInfo**
> containerArchiveInfo(id, path)

Get information about files in a container

    Get information about files in a container in the form of a stat structure

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **path** | **String**| Resource path in the container | [default to null] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerAttach"></a>
# **containerAttach**
> File containerAttach(id, detachKeys, logs, stream, stdin, stdout, stderr)

Attach to a container

    Attach to a container to read its output or send it input

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **detachKeys** | **String**| Override the key sequence for detaching a container | [optional] [default to null] |
| **logs** | **Boolean**| Return logs | [optional] [default to false] |
| **stream** | **Boolean**| Return stream | [optional] [default to false] |
| **stdin** | **Boolean**| Attach to stdin | [optional] [default to false] |
| **stdout** | **Boolean**| Attach to stdout | [optional] [default to false] |
| **stderr** | **Boolean**| Attach to stderr | [optional] [default to false] |

### Return type

**File**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream, application/json

<a name="containerCreate"></a>
# **containerCreate**
> CreateContainerResponse containerCreate(CreateContainerRequest, name)

Create a container

    Create a new container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **CreateContainerRequest** | [**CreateContainerRequest**](../Models/CreateContainerRequest.md)| Container configuration | |
| **name** | **String**| Container name | [optional] [default to null] |

### Return type

[**CreateContainerResponse**](../Models/CreateContainerResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="containerDelete"></a>
# **containerDelete**
> containerDelete(id, v, force)

Remove a container

    Remove a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **v** | **Boolean**| Remove anonymous volumes associated with the container | [optional] [default to false] |
| **force** | **Boolean**| Force the removal of a running container | [optional] [default to false] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerExec"></a>
# **containerExec**
> ContainerExec_201_response containerExec(id, ExecConfig)

Create an exec instance

    Run a command inside a running container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **ExecConfig** | [**ExecConfig**](../Models/ExecConfig.md)| Exec configuration | |

### Return type

[**ContainerExec_201_response**](../Models/ContainerExec_201_response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="containerInspect"></a>
# **containerInspect**
> ContainerInspectResponse containerInspect(id, size)

Inspect a container

    Return low-level information about a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **size** | **Boolean**| Return the size of container as fields SizeRw and SizeRootFs | [optional] [default to false] |

### Return type

[**ContainerInspectResponse**](../Models/ContainerInspectResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerKill"></a>
# **containerKill**
> containerKill(id, signal)

Kill a container

    Kill a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **signal** | **String**| Signal to send to the container | [optional] [default to KILL] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerList"></a>
# **containerList**
> List containerList(all, limit, size, filters)

List containers

    Returns a list of containers

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **all** | **Boolean**| Show all containers (default shows just running) | [optional] [default to false] |
| **limit** | **Integer**| Show limit last created containers | [optional] [default to null] |
| **size** | **Boolean**| Show the containers sizes | [optional] [default to false] |
| **filters** | **String**| Filter output based on conditions provided | [optional] [default to null] |

### Return type

[**List**](../Models/ContainerSummary.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerLogs"></a>
# **containerLogs**
> File containerLogs(id, follow, stdout, stderr, since, until, timestamps, tail)

Get container logs

    Get stdout and stderr logs from a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **follow** | **Boolean**| Follow log output | [optional] [default to false] |
| **stdout** | **Boolean**| Show stdout log | [optional] [default to false] |
| **stderr** | **Boolean**| Show stderr log | [optional] [default to false] |
| **since** | **String**| Show logs since timestamp | [optional] [default to null] |
| **until** | **String**| Show logs before timestamp | [optional] [default to null] |
| **timestamps** | **Boolean**| Show timestamps | [optional] [default to false] |
| **tail** | **String**| Number of lines to show from the end of the logs | [optional] [default to all] |

### Return type

**File**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream, application/json

<a name="containerPause"></a>
# **containerPause**
> containerPause(id)

Pause a container

    Pause a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerRemove"></a>
# **containerRemove**
> containerRemove(id, v, force)

Remove a container

    Remove a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **v** | **Boolean**| Remove anonymous volumes associated with the container | [optional] [default to false] |
| **force** | **Boolean**| Force the removal of a running container | [optional] [default to false] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerRename"></a>
# **containerRename**
> containerRename(id, name)

Rename a container

    Rename a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **name** | **String**| New name for the container | [default to null] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerRestart"></a>
# **containerRestart**
> containerRestart(id, t)

Restart a container

    Restart a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **t** | **Integer**| Seconds to wait before killing the container | [optional] [default to 10] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerStart"></a>
# **containerStart**
> containerStart(id, detachKeys)

Start a container

    Start a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **detachKeys** | **String**| Override the key sequence for detaching a container | [optional] [default to null] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerStats"></a>
# **containerStats**
> Object containerStats(id, stream)

Get container stats based on resource usage

    This endpoint returns a live stream of a container&#39;s resource usage statistics

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **stream** | **Boolean**| Stream statistics | [optional] [default to true] |

### Return type

**Object**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerStop"></a>
# **containerStop**
> containerStop(id, t)

Stop a container

    Stop a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **t** | **Integer**| Seconds to wait before killing the container | [optional] [default to 10] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerUnpause"></a>
# **containerUnpause**
> containerUnpause(id)

Unpause a container

    Unpause a container

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="containerWait"></a>
# **containerWait**
> ContainerWaitResponse containerWait(id, condition)

Wait for a container

    Block until a container stops, then returns the exit code

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Container ID or name | [default to null] |
| **condition** | **String**| Wait until a container state reaches the given condition | [optional] [default to not-running] [enum: not-running, next-exit, removed] |

### Return type

[**ContainerWaitResponse**](../Models/ContainerWaitResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="distributionInspect"></a>
# **distributionInspect**
> DistributionInspect distributionInspect(name)

Get image distribution information

    Return image digest and platform information

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | **String**| Image name or ID | [default to null] |

### Return type

[**DistributionInspect**](../Models/DistributionInspect.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="execInspect"></a>
# **execInspect**
> ExecInspectResponse execInspect(id)

Inspect an exec instance

    Return low-level information about an exec instance

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Exec instance ID | [default to null] |

### Return type

[**ExecInspectResponse**](../Models/ExecInspectResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="execResize"></a>
# **execResize**
> execResize(id, h, w)

Resize an exec instance

    Resize the TTY for an exec instance

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Exec instance ID | [default to null] |
| **h** | **Integer**| Height of the TTY session in characters | [optional] [default to null] |
| **w** | **Integer**| Width of the TTY session in characters | [optional] [default to null] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="execStart"></a>
# **execStart**
> File execStart(id, ExecStartConfig)

Start an exec instance

    Starts a previously created exec instance

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Exec instance ID | [default to null] |
| **ExecStartConfig** | [**ExecStartConfig**](../Models/ExecStartConfig.md)| Exec start configuration | |

### Return type

**File**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/octet-stream, application/json

<a name="imageBuild"></a>
# **imageBuild**
> BuildResult imageBuild(body, dockerfile, t, q, nocache, rm, forcerm, platform, target, buildargs, labels, cachefrom, networkmode, output)

Build an image

    Build an image from a Dockerfile

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **body** | **File**| A tar archive compressed with one of the following algorithms gzip, bzip2, xz | |
| **dockerfile** | **String**| Name of the Dockerfile | [optional] [default to Dockerfile] |
| **t** | **String**| Name and optionally a tag in the &#39;name:tag&#39; format | [optional] [default to null] |
| **q** | **Boolean**| Suppress verbose build output | [optional] [default to false] |
| **nocache** | **Boolean**| Do not use the cache when building the image | [optional] [default to false] |
| **rm** | **Boolean**| Remove intermediate containers after a successful build | [optional] [default to true] |
| **forcerm** | **Boolean**| Always remove intermediate containers, even upon failure | [optional] [default to false] |
| **platform** | **String**| Platform in the format os[/arch[/variant]] | [optional] [default to null] |
| **target** | **String**| Set the target build stage to build | [optional] [default to null] |
| **buildargs** | **String**| Build-time variables as JSON | [optional] [default to null] |
| **labels** | **String**| Set metadata for an image as JSON | [optional] [default to null] |
| **cachefrom** | **String**| Images to consider as cache sources as JSON | [optional] [default to null] |
| **networkmode** | **String**| Set the networking mode for the RUN instructions during build | [optional] [default to null] |
| **output** | **String**| Output destination | [optional] [default to null] |

### Return type

[**BuildResult**](../Models/BuildResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/octet-stream
- **Accept**: application/json

<a name="imageCreate"></a>
# **imageCreate**
> Object imageCreate(fromImage, tag, platform)

Create an image

    Create an image by pulling it from a registry

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **fromImage** | **String**| Name of the image to pull | [default to null] |
| **tag** | **String**| Tag of the image to pull | [optional] [default to latest] |
| **platform** | **String**| Platform in the format os[/arch[/variant]] | [optional] [default to null] |

### Return type

**Object**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="imageDelete"></a>
# **imageDelete**
> List imageDelete(name, force, noprune)

Remove an image

    Remove an image

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | **String**| Image name or ID | [default to null] |
| **force** | **Boolean**| Force removal of the image | [optional] [default to false] |
| **noprune** | **Boolean**| Do not delete untagged parent images | [optional] [default to false] |

### Return type

[**List**](../Models/ImageDelete_200_response_inner.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="imageInspect"></a>
# **imageInspect**
> ImageInspect imageInspect(name)

Inspect an image

    Return low-level information about an image

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | **String**| Image name or ID | [default to null] |

### Return type

[**ImageInspect**](../Models/ImageInspect.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="imageList"></a>
# **imageList**
> List imageList(all, filters, digests)

List images

    Returns a list of images on the server

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **all** | **Boolean**| Show all images (default hides intermediate images) | [optional] [default to false] |
| **filters** | **String**| Filter output based on conditions provided | [optional] [default to null] |
| **digests** | **Boolean**| Show digest information | [optional] [default to false] |

### Return type

[**List**](../Models/ImageSummary.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="imageLoad"></a>
# **imageLoad**
> ImageLoadResponse imageLoad(body, quiet)

Load an image

    Load an image from a tar archive

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **body** | **File**| Tar archive containing the image | |
| **quiet** | **Boolean**| Suppress progress details during load | [optional] [default to false] |

### Return type

[**ImageLoadResponse**](../Models/ImageLoadResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/x-tar
- **Accept**: application/json

<a name="imagePush"></a>
# **imagePush**
> Object imagePush(name, tag)

Push an image

    Push an image to a registry

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | **String**| Image name or ID | [default to null] |
| **tag** | **String**| Tag of the image to pull | [optional] [default to latest] |

### Return type

**Object**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="imageTag"></a>
# **imageTag**
> imageTag(name, repo, tag)

Tag an image

    Tag an image so that it becomes part of a repository

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | **String**| Image name or ID | [default to null] |
| **repo** | **String**| The repository to tag in | [default to null] |
| **tag** | **String**| Tag of the image to pull | [optional] [default to latest] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="networkConnect"></a>
# **networkConnect**
> networkConnect(id, NetworkConnectRequest)

Connect a container to a network

    Connect a container to a network

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Network ID or name | [default to null] |
| **NetworkConnectRequest** | [**NetworkConnectRequest**](../Models/NetworkConnectRequest.md)| Container configuration | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="networkCreate"></a>
# **networkCreate**
> NetworkCreateResponse networkCreate(NetworkCreateRequest)

Create a network

    Create a network

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **NetworkCreateRequest** | [**NetworkCreateRequest**](../Models/NetworkCreateRequest.md)| Network configuration | |

### Return type

[**NetworkCreateResponse**](../Models/NetworkCreateResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="networkDelete"></a>
# **networkDelete**
> networkDelete(id)

Remove a network

    Remove a network

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Network ID or name | [default to null] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="networkInspect"></a>
# **networkInspect**
> Network networkInspect(id)

Inspect a network

    Return low-level information about a network

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Network ID or name | [default to null] |

### Return type

[**Network**](../Models/Network.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="networkList"></a>
# **networkList**
> List networkList(filters)

List networks

    Returns a list of networks

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **filters** | **String**| Filter output based on conditions provided | [optional] [default to null] |

### Return type

[**List**](../Models/Network.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="systemAuth"></a>
# **systemAuth**
> AuthResponse systemAuth(AuthConfig)

Check auth configuration

    Check auth configuration

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **AuthConfig** | [**AuthConfig**](../Models/AuthConfig.md)| Authentication configuration | |

### Return type

[**AuthResponse**](../Models/AuthResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="systemEvents"></a>
# **systemEvents**
> SystemEvents_200_response systemEvents(since, until, filters)

Monitor events

    Stream real-time events from the server

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **since** | **String**| Show events created since this timestamp | [optional] [default to null] |
| **until** | **String**| Show events created until this timestamp | [optional] [default to null] |
| **filters** | **String**| Filter output based on conditions provided | [optional] [default to null] |

### Return type

[**SystemEvents_200_response**](../Models/SystemEvents_200_response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="systemInfo"></a>
# **systemInfo**
> SystemInfo systemInfo()

Get system information

    Return system information

### Parameters
This endpoint does not need any parameter.

### Return type

[**SystemInfo**](../Models/SystemInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="systemPing"></a>
# **systemPing**
> String systemPing()

Ping

    Ping the server

### Parameters
This endpoint does not need any parameter.

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain, application/json

<a name="systemVersion"></a>
# **systemVersion**
> VersionResponse systemVersion()

Get version

    Return version information

### Parameters
This endpoint does not need any parameter.

### Return type

[**VersionResponse**](../Models/VersionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="volumeCreate"></a>
# **volumeCreate**
> Volume volumeCreate(VolumeCreateRequest)

Create a volume

    Create a volume

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **VolumeCreateRequest** | [**VolumeCreateRequest**](../Models/VolumeCreateRequest.md)| Volume configuration | |

### Return type

[**Volume**](../Models/Volume.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="volumeDelete"></a>
# **volumeDelete**
> volumeDelete(name, force)

Remove a volume

    Remove a volume

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | **String**| Volume name or ID | [default to null] |
| **force** | **Boolean**| Force the removal of the volume | [optional] [default to false] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="volumeInspect"></a>
# **volumeInspect**
> Volume volumeInspect(name)

Inspect a volume

    Return low-level information about a volume

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | **String**| Volume name or ID | [default to null] |

### Return type

[**Volume**](../Models/Volume.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="volumeList"></a>
# **volumeList**
> VolumeListResponse volumeList(filters)

List volumes

    Returns a list of volumes

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **filters** | **String**| Filter output based on conditions provided | [optional] [default to null] |

### Return type

[**VolumeListResponse**](../Models/VolumeListResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

