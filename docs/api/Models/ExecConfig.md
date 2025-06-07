# ExecConfig
## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
| **AttachStdin** | **Boolean** | Attach to stdin | [optional] [default to false] |
| **AttachStdout** | **Boolean** | Attach to stdout | [optional] [default to true] |
| **AttachStderr** | **Boolean** | Attach to stderr | [optional] [default to true] |
| **DetachKeys** | **String** | Override the key sequence for detaching a container | [optional] [default to null] |
| **Tty** | **Boolean** | Allocate a pseudo-TTY | [optional] [default to false] |
| **Cmd** | **List** | Command to run | [optional] [default to null] |
| **Env** | **List** | Environment variables | [optional] [default to null] |
| **WorkingDir** | **String** | Working directory | [optional] [default to null] |
| **Privileged** | **Boolean** | Give extended privileges to the command | [optional] [default to false] |
| **User** | **String** | User that will run the command | [optional] [default to null] |

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

