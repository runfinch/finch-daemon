# NetworkCreateRequest
## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
| **Name** | **String** | Network name | [default to null] |
| **CheckDuplicate** | **Boolean** | Check for duplicate networks | [optional] [default to null] |
| **Driver** | **String** | Network driver | [optional] [default to bridge] |
| **Internal** | **Boolean** | Whether the network is internal | [optional] [default to null] |
| **Attachable** | **Boolean** | Whether the network is attachable | [optional] [default to null] |
| **Ingress** | **Boolean** | Whether the network is ingress | [optional] [default to null] |
| **IPAM** | [**NetworkCreateRequest_IPAM**](NetworkCreateRequest_IPAM.md) |  | [optional] [default to null] |
| **EnableIPv6** | **Boolean** | Whether IPv6 is enabled on the network | [optional] [default to null] |
| **Options** | **Map** | Network options | [optional] [default to null] |
| **Labels** | **Map** | Network labels | [optional] [default to null] |

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

