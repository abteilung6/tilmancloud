# DefaultApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createNode**](#createnode) | **POST** /nodes | Create a new node|
|[**deleteNode**](#deletenode) | **DELETE** /nodes/{nodeId} | Delete a node|
|[**health**](#health) | **GET** /health | Health check|
|[**listImages**](#listimages) | **GET** /images | List all images|
|[**listNodes**](#listnodes) | **GET** /nodes | List all nodes|

# **createNode**
> Node createNode()

Creates a new node.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@tilmancloud/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.createNode();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Node**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Node created successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteNode**
> deleteNode()

Deletes (terminates) a node by its ID

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@tilmancloud/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let nodeId: string; //The ID of the node to delete (default to undefined)

const { status, data } = await apiInstance.deleteNode(
    nodeId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **nodeId** | [**string**] | The ID of the node to delete | defaults to undefined|


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**204** | Node deleted successfully |  -  |
|**404** | Node not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **health**
> Health health()

Returns the health status of the API

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@tilmancloud/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.health();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Health**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | API is healthy |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listImages**
> Array<Image> listImages()

Returns a list of all AMIs (Amazon Machine Images)

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@tilmancloud/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.listImages();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<Image>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of images |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listNodes**
> Array<Node> listNodes()

Returns a list of all nodes (EC2 instances)

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@tilmancloud/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.listNodes();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<Node>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of nodes |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

