# SystemInfo
## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
| **ID** | **String** | Unique identifier of the daemon | [optional] [default to null] |
| **Containers** | **Integer** | Total number of containers | [optional] [default to null] |
| **ContainersRunning** | **Integer** | Number of containers running | [optional] [default to null] |
| **ContainersPaused** | **Integer** | Number of containers paused | [optional] [default to null] |
| **ContainersStopped** | **Integer** | Number of containers stopped | [optional] [default to null] |
| **Images** | **Integer** | Number of images | [optional] [default to null] |
| **Driver** | **String** | Storage driver | [optional] [default to null] |
| **DriverStatus** | [**List**](array.md) | Storage driver status | [optional] [default to null] |
| **SystemStatus** | [**List**](array.md) | System status | [optional] [default to null] |
| **Plugins** | [**SystemInfo_Plugins**](SystemInfo_Plugins.md) |  | [optional] [default to null] |
| **MemoryLimit** | **Boolean** | Whether memory limit support is enabled | [optional] [default to null] |
| **SwapLimit** | **Boolean** | Whether swap limit support is enabled | [optional] [default to null] |
| **KernelMemory** | **Boolean** | Whether kernel memory limit support is enabled | [optional] [default to null] |
| **KernelMemoryTCP** | **Boolean** | Whether kernel memory TCP limit support is enabled | [optional] [default to null] |
| **CpuCfsPeriod** | **Boolean** | Whether CPU CFS period support is enabled | [optional] [default to null] |
| **CpuCfsQuota** | **Boolean** | Whether CPU CFS quota support is enabled | [optional] [default to null] |
| **CPUShares** | **Boolean** | Whether CPU shares support is enabled | [optional] [default to null] |
| **CPUSet** | **Boolean** | Whether CPUSet support is enabled | [optional] [default to null] |
| **PidsLimit** | **Boolean** | Whether PIDs limit support is enabled | [optional] [default to null] |
| **IPv4Forwarding** | **Boolean** | Whether IPv4 forwarding is enabled | [optional] [default to null] |
| **BridgeNfIptables** | **Boolean** | Whether bridge netfilter iptables is enabled | [optional] [default to null] |
| **BridgeNfIp6tables** | **Boolean** | Whether bridge netfilter ip6tables is enabled | [optional] [default to null] |
| **Debug** | **Boolean** | Whether debug mode is enabled | [optional] [default to null] |
| **NFd** | **Integer** | Number of file descriptors | [optional] [default to null] |
| **OomKillDisable** | **Boolean** | Whether OOM kill is disabled | [optional] [default to null] |
| **NGoroutines** | **Integer** | Number of goroutines | [optional] [default to null] |
| **SystemTime** | **Date** | Current system time | [optional] [default to null] |
| **LoggingDriver** | **String** | Logging driver | [optional] [default to null] |
| **CgroupDriver** | **String** | Cgroup driver | [optional] [default to null] |
| **CgroupVersion** | **String** | Cgroup version | [optional] [default to null] |
| **NEventsListener** | **Integer** | Number of events listeners | [optional] [default to null] |
| **KernelVersion** | **String** | Kernel version | [optional] [default to null] |
| **OperatingSystem** | **String** | Operating system | [optional] [default to null] |
| **OSVersion** | **String** | Operating system version | [optional] [default to null] |
| **OSType** | **String** | Operating system type | [optional] [default to null] |
| **Architecture** | **String** | Hardware architecture | [optional] [default to null] |
| **IndexServerAddress** | **String** | Index server address | [optional] [default to null] |
| **RegistryConfig** | [**SystemInfo_RegistryConfig**](SystemInfo_RegistryConfig.md) |  | [optional] [default to null] |
| **NCPU** | **Integer** | Number of CPUs | [optional] [default to null] |
| **MemTotal** | **Long** | Total memory | [optional] [default to null] |
| **GenericResources** | [**List**](SystemInfo_GenericResources_inner.md) | Generic resources | [optional] [default to null] |
| **DockerRootDir** | **String** | Docker root directory | [optional] [default to null] |
| **HttpProxy** | **String** | HTTP proxy | [optional] [default to null] |
| **HttpsProxy** | **String** | HTTPS proxy | [optional] [default to null] |
| **NoProxy** | **String** | No proxy | [optional] [default to null] |
| **Name** | **String** | Name | [optional] [default to null] |
| **Labels** | **List** | Labels | [optional] [default to null] |
| **ExperimentalBuild** | **Boolean** | Whether experimental build is enabled | [optional] [default to null] |
| **ServerVersion** | **String** | Server version | [optional] [default to null] |
| **Runtimes** | [**Map**](SystemInfo_Runtimes_value.md) | Runtimes | [optional] [default to null] |
| **DefaultRuntime** | **String** | Default runtime | [optional] [default to null] |
| **Swarm** | [**SystemInfo_Swarm**](SystemInfo_Swarm.md) |  | [optional] [default to null] |
| **LiveRestoreEnabled** | **Boolean** | Whether live restore is enabled | [optional] [default to null] |
| **Isolation** | **String** | Isolation | [optional] [default to null] |
| **InitBinary** | **String** | Init binary | [optional] [default to null] |
| **ContainerdCommit** | [**SystemInfo_ContainerdCommit**](SystemInfo_ContainerdCommit.md) |  | [optional] [default to null] |
| **RuncCommit** | [**SystemInfo_RuncCommit**](SystemInfo_RuncCommit.md) |  | [optional] [default to null] |
| **InitCommit** | [**SystemInfo_InitCommit**](SystemInfo_InitCommit.md) |  | [optional] [default to null] |
| **SecurityOptions** | **List** | Security options | [optional] [default to null] |
| **ProductLicense** | **String** | Product license | [optional] [default to null] |
| **DefaultAddressPools** | [**List**](SystemInfo_DefaultAddressPools_inner.md) | Default address pools | [optional] [default to null] |
| **Warnings** | **List** | Warnings | [optional] [default to null] |

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

