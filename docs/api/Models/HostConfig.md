# HostConfig
## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
| **Binds** | **List** | List of volume bindings for this container (e.g., [\&quot;/host:/container:ro\&quot;]) | [optional] [default to null] |
| **ContainerIDFile** | **String** | File path where the container ID is written | [optional] [default to null] |
| **LogConfig** | [**HostConfig_LogConfig**](HostConfig_LogConfig.md) |  | [optional] [default to null] |
| **NetworkMode** | **String** | Network mode to use for the container (e.g., \&quot;bridge\&quot;, \&quot;host\&quot;, \&quot;none\&quot;) | [optional] [default to bridge] |
| **PortBindings** | [**Map**](array.md) | Port mapping between the exposed port (container) and the host | [optional] [default to null] |
| **RestartPolicy** | [**HostConfig_RestartPolicy**](HostConfig_RestartPolicy.md) |  | [optional] [default to null] |
| **AutoRemove** | **Boolean** | Automatically remove the container when it exits | [optional] [default to false] |
| **VolumesFrom** | **List** | List of volumes to take from other containers | [optional] [default to null] |
| **CapAdd** | **List** | List of kernel capabilities to add to the container | [optional] [default to null] |
| **CapDrop** | **List** | List of kernel capabilities to remove from the container | [optional] [default to null] |
| **CgroupnsMode** | **String** | Cgroup namespace mode to use for the container (\&quot;host\&quot; or \&quot;private\&quot;) | [optional] [default to null] |
| **Dns** | **List** | List of DNS servers for the container | [optional] [default to null] |
| **DnsOptions** | **List** | List of DNS options for the container | [optional] [default to null] |
| **DnsSearch** | **List** | List of DNS search domains for the container | [optional] [default to null] |
| **ExtraHosts** | **List** | List of hostnames/IP mappings to add to /etc/hosts | [optional] [default to null] |
| **GroupAdd** | **List** | List of additional groups that the container process will run as | [optional] [default to null] |
| **IpcMode** | **String** | IPC namespace to use for the container | [optional] [default to null] |
| **OomKillDisable** | **Boolean** | Whether to disable OOM Killer for the container | [optional] [default to false] |
| **PidMode** | **String** | PID namespace to use for the container | [optional] [default to null] |
| **Privileged** | **Boolean** | Give extended privileges to this container | [optional] [default to false] |
| **ReadonlyRootfs** | **Boolean** | Mount the container&#39;s root filesystem as read only | [optional] [default to false] |
| **SecurityOpt** | **List** | List of security options for the container | [optional] [default to null] |
| **Tmpfs** | **Map** | Temporary filesystems to mount | [optional] [default to null] |
| **UTSMode** | **String** | UTS namespace to use for the container | [optional] [default to null] |
| **ShmSize** | **Long** | Size of /dev/shm in bytes | [optional] [default to null] |
| **Sysctls** | **Map** | Kernel parameters to set in the container | [optional] [default to null] |
| **Runtime** | **String** | Runtime to use for this container | [optional] [default to null] |
| **CpuShares** | **Long** | CPU shares (relative weight vs. other containers) | [optional] [default to null] |
| **CpuPeriod** | **Long** | CPU CFS (Completely Fair Scheduler) period | [optional] [default to null] |
| **CpuQuota** | **Long** | CPU CFS (Completely Fair Scheduler) quota | [optional] [default to null] |
| **CpusetCpus** | **String** | CPUs in which to allow execution (e.g., \&quot;0-3\&quot;, \&quot;0,1\&quot;) | [optional] [default to null] |
| **CpusetMems** | **String** | Memory nodes (MEMs) in which to allow execution (e.g., \&quot;0-3\&quot;, \&quot;0,1\&quot;) | [optional] [default to null] |
| **Memory** | **Long** | Memory limit in bytes | [optional] [default to null] |
| **MemoryReservation** | **Long** | Memory soft limit in bytes | [optional] [default to null] |
| **MemorySwap** | **Long** | Total memory usage (memory + swap); set &#x60;-1&#x60; to enable unlimited swap | [optional] [default to null] |
| **MemorySwappiness** | **Integer** | Tune container memory swappiness (0 to 100) | [optional] [default to null] |
| **BlkioWeight** | **Integer** | Block IO weight (relative weight vs. other containers). The weight is a value between 10 and 1000 that affects the scheduling priority for block IO operations. | [optional] [default to null] |
| **BlkioWeightDevice** | [**List**](HostConfig_BlkioWeightDevice_inner.md) | Block IO weight for specific devices. Each item in the array specifies a device path and the IO weight for that device. | [optional] [default to null] |
| **BlkioDeviceReadBps** | [**List**](HostConfig_BlkioDeviceReadBps_inner.md) | Limit read rate from a device (bytes per second). Each item in the array specifies a device path and the rate limit for that device. | [optional] [default to null] |
| **BlkioDeviceWriteBps** | [**List**](HostConfig_BlkioDeviceReadBps_inner.md) | Limit write rate to a device (bytes per second). Each item in the array specifies a device path and the rate limit for that device. | [optional] [default to null] |
| **BlkioDeviceReadIOps** | [**List**](HostConfig_BlkioDeviceReadIOps_inner.md) | Limit read rate from a device (IO per second). Each item in the array specifies a device path and the IO rate limit for that device. | [optional] [default to null] |
| **BlkioDeviceWriteIOps** | [**List**](HostConfig_BlkioDeviceReadIOps_inner.md) | Limit write rate to a device (IO per second). Each item in the array specifies a device path and the IO rate limit for that device. | [optional] [default to null] |
| **Devices** | [**List**](HostConfig_Devices_inner.md) | Expose host devices to the container. Each item in the array specifies a device mapping between the host and the container. | [optional] [default to null] |
| **PidsLimit** | **Long** | Tune container pids limit (set -1 for unlimited). This limits the number of processes that can run inside the container. | [optional] [default to null] |

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

