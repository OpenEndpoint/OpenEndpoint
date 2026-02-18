# OpenEndpoint Error Codes

This document describes the error codes returned by the OpenEndpoint S3-compatible API.

## Common Error Response Structure

All error responses follow the AWS S3 error format:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Error>
  <Code>NoSuchBucket</Code>
  <Message>The specified bucket does not exist.</Message>
  <BucketName>my-bucket</BucketName>
  <RequestId>1234567890ABCDEF</RequestId>
  <HostId>some-host-id</HostId>
</Error>
```

## Error Codes

### Bucket Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `NoSuchBucket` | 404 | The specified bucket does not exist. |
| `BucketAlreadyExists` | 409 | The requested bucket name is not available. |
| `BucketNotEmpty` | 409 | The bucket you tried to delete is not empty. |
| `InvalidBucketName` | 400 | The specified bucket name is invalid. |
| `TooManyBuckets` | 400 | You have attempted to create more buckets than allowed. |

### Object Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `NoSuchKey` | 404 | The specified key does not exist. |
| `InvalidObjectName` | 400 | The specified object name is invalid. |
| `ObjectLockNotEnabled` | 400 | Object lock is not enabled on this bucket. |
| `InvalidObjectLockState` | 400 | The object lock state is invalid for this operation. |

### Authentication Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `AccessDenied` | 403 | Access denied. |
| `InvalidAccessKeyId` | 403 | The Access Key Id provided does not exist in our records. |
| `SignatureDoesNotMatch` | 403 | The request signature we calculated does not match the signature you provided. |
| `ExpiredPresignedRequest` | 403 | The presigned URL has expired. |
| `MissingSecurityHeader` | 400 | The security header is required but not provided. |
| `InvalidSecurity` | 403 | The provided security credentials are not valid. |

### Multipart Upload Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `NoSuchUpload` | 404 | The specified multipart upload does not exist. |
| `InvalidPart` | 400 | One or more of the specified parts could not be found. |
| `InvalidPartOrder` | 400 | The list of parts was not in order. |

### Lifecycle Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `NoSuchLifecycle` | 404 | The lifecycle configuration does not exist. |
| `MalformedXML` | 400 | The XML provided was not well-formed or did not validate against schema. |

### General Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `InternalError` | 500 | An internal error occurred. Please retry. |
| `InvalidURI` | 400 | The specified URI is invalid. |
| `MethodNotAllowed` | 405 | The specified method is not allowed. |
| `RequestTimeout` | 400 | The request timed out. |
| `NotImplemented` | 501 | The requested feature is not implemented. |

## Error Handling Best Practices

1. **Check HTTP Status Code**: Most errors return a 4xx status code. 403 typically means authentication/authorization failure, 404 means resource not found.

2. **Parse Error Code**: Always check the `<Code>` element to determine the specific error type.

3. **Retry Logic**: For `InternalError` (500), implement exponential backoff retry logic.

4. **Presigned URLs**: Check `ExpiredPresignedRequest` to regenerate expired presigned URLs.

5. **Bucket Names**: Validate bucket names before creation to avoid `InvalidBucketName` errors.

## SDK Compatibility

OpenEndpoint aims to be compatible with AWS S3 SDKs. The error codes are modeled after AWS S3 error codes to ensure SDK compatibility.

### Supported SDKs

- AWS SDK for Go
- AWS SDK for Python (Boto3)
- AWS SDK for Java
- AWS CLI

## Debugging Tips

1. Enable debug logging to see full request/response details
2. Check the `<RequestId>` in error responses for support tickets
3. Verify authentication credentials and bucket/object names
4. Ensure proper permissions are set via bucket policies or IAM
