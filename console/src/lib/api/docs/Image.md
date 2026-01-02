# Image


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | AMI ID | [default to undefined]
**name** | **string** | AMI name | [optional] [default to undefined]
**state** | **string** | AMI state | [default to undefined]
**imageId** | **string** | Content-based image identifier (from ImageID tag) | [optional] [default to undefined]
**creationDate** | **string** | AMI creation date (ISO 8601) | [default to undefined]
**description** | **string** | AMI description | [optional] [default to undefined]
**snapshotId** | **string** | Source snapshot ID (from SnapshotID tag) | [optional] [default to undefined]
**architecture** | **string** | Architecture type | [default to undefined]
**virtualizationType** | **string** | Virtualization type | [default to undefined]

## Example

```typescript
import { Image } from '@tilmancloud/api-client';

const instance: Image = {
    id,
    name,
    state,
    imageId,
    creationDate,
    description,
    snapshotId,
    architecture,
    virtualizationType,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
