# Network
## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
| **Name** | **String** | Network name | [optional] [default to null] |
| **Id** | **String** | Network ID | [optional] [default to null] |
| **Created** | **Date** | When the network was created | [optional] [default to null] |
| **Scope** | **String** | Network scope | [optional] [default to null] |
| **Driver** | **String** | Network driver | [optional] [default to null] |
| **EnableIPv6** | **Boolean** | Whether IPv6 is enabled on the network | [optional] [default to null] |
| **IPAM** | [**Network_IPAM**](Network_IPAM.md) |  | [optional] [default to null] |
| **Internal** | **Boolean** | Whether the network is internal | [optional] [default to null] |
| **Attachable** | **Boolean** | Whether the network is attachable | [optional] [default to null] |
| **Ingress** | **Boolean** | Whether the network is ingress | [optional] [default to null] |
| **ConfigFrom** | [**Network_ConfigFrom**](Network_ConfigFrom.md) |  | [optional] [default to null] |
| **ConfigOnly** | **Boolean** | Whether the network is a configuration only network | [optional] [default to null] |
| **Containers** | [**Map**](Network_Containers_value.md) | Containers connected to the network | [optional] [default to null] |
| **Options** | **Map** | Network options | [optional] [default to null] |
| **Labels** | **Map** | Network labels | [optional] [default to null] |

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

