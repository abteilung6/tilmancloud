# Node


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** | Node name (EC2 instance ID) | [default to undefined]
**state** | **string** | Current node state | [optional] [default to undefined]
**instanceType** | **string** | EC2 instance type | [optional] [default to undefined]
**publicIp** | **string** | Public IP address | [optional] [default to undefined]
**privateIp** | **string** | Private IP address | [optional] [default to undefined]

## Example

```typescript
import { Node } from '@tilmancloud/api-client';

const instance: Node = {
    name,
    state,
    instanceType,
    publicIp,
    privateIp,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
