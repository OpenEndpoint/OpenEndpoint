package api

import (
	"errors"
	"net/http"
	"testing"
)

func TestS3ErrorCode(t *testing.T) {
	tests := []struct {
		err        error
		wantCode   string
		wantStatus int
	}{
		{ErrInternal, "InternalError", http.StatusInternalServerError},
		{ErrInvalidURI, "InvalidURI", http.StatusBadRequest},
		{ErrNoSuchBucket, "NoSuchBucket", http.StatusNotFound},
		{ErrNoSuchKey, "NoSuchKey", http.StatusNotFound},
		{ErrBucketExists, "BucketAlreadyExists", http.StatusConflict},
		{ErrBucketNotEmpty, "BucketNotEmpty", http.StatusConflict},
		{ErrInvalidBucketName, "InvalidBucketName", http.StatusBadRequest},
		{ErrInvalidObjectName, "InvalidObjectName", http.StatusBadRequest},
		{ErrAccessDenied, "AccessDenied", http.StatusForbidden},
		{ErrSignatureDoesNotMatch, "SignatureDoesNotMatch", http.StatusForbidden},
		{ErrInvalidAccessKeyID, "InvalidAccessKeyId", http.StatusForbidden},
		{ErrExpiredPresignedRequest, "ExpiredPresignedRequest", http.StatusForbidden},
		{ErrMissingSecurityHeader, "MissingSecurityHeader", http.StatusBadRequest},
		{ErrInvalidSecurity, "InvalidSecurity", http.StatusForbidden},
		{ErrRequestTimeout, "RequestTimeout", http.StatusBadRequest},
		{ErrMalformedXML, "MalformedXML", http.StatusBadRequest},
		{ErrMethodNotAllowed, "MethodNotAllowed", http.StatusMethodNotAllowed},
		{ErrTooManyBuckets, "TooManyBuckets", http.StatusBadRequest},
		{ErrNoSuchUpload, "NoSuchUpload", http.StatusNotFound},
		{ErrInvalidPart, "InvalidPart", http.StatusBadRequest},
		{ErrInvalidPartOrder, "InvalidPartOrder", http.StatusBadRequest},
		{ErrNoSuchLifecycle, "NoSuchLifecycle", http.StatusNotFound},
		{ErrBucketNotEmpty, "BucketNotEmpty", http.StatusConflict},
		{ErrObjectLockNotEnabled, "ObjectLockNotEnabled", http.StatusBadRequest},
		{ErrInvalidObjectLockState, "InvalidObjectLockState", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.wantCode, func(t *testing.T) {
			s3Err, ok := tt.err.(S3Error)
			if !ok {
				t.Errorf("error %v is not an S3Error", tt.err)
				return
			}

			if s3Err.Code() != tt.wantCode {
				t.Errorf("Code() = %v, want %v", s3Err.Code(), tt.wantCode)
			}

			if s3Err.StatusCode() != tt.wantStatus {
				t.Errorf("StatusCode() = %v, want %v", s3Err.StatusCode(), tt.wantStatus)
			}
		})
	}
}

func TestS3ErrorMessage(t *testing.T) {
	tests := []struct {
		err      error
		wantMsg  string
	}{
		{ErrInternal, "The request processing has failed due to some unknown error."},
		{ErrInvalidURI, "The specified URI is invalid."},
		{ErrNoSuchBucket, "The specified bucket does not exist."},
		{ErrNoSuchKey, "The specified key does not exist."},
		{ErrBucketExists, "The requested bucket name is not available."},
		{ErrBucketNotEmpty, "The bucket you tried to delete is not empty."},
		{ErrInvalidBucketName, "The specified bucket name is invalid."},
		{ErrInvalidObjectName, "The specified object name is invalid."},
		{ErrAccessDenied, "Access Denied."},
		{ErrSignatureDoesNotMatch, "The request signature we calculated does not match the signature you provided."},
		{ErrInvalidAccessKeyID, "The Access Key Id you provided does not exist in our records."},
		{ErrExpiredPresignedRequest, "The request has expired."},
		{ErrMissingSecurityHeader, "The security header is required."},
		{ErrMethodNotAllowed, "The specified method is not allowed against this resource."},
		{ErrTooManyBuckets, "You have attempted to create more buckets than allowed."},
		{ErrNoSuchUpload, "The specified multipart upload does not exist."},
		{ErrInvalidPart, "One or more of the specified parts could not be found."},
		{ErrInvalidPartOrder, "The list of parts was not in order."},
	}

	for _, tt := range tests {
		t.Run(tt.wantMsg[:20], func(t *testing.T) {
			s3Err, ok := tt.err.(S3Error)
			if !ok {
				t.Errorf("error %v is not an S3Error", tt.err)
				return
			}

			if s3Err.Message() != tt.wantMsg {
				t.Errorf("Message() = %v, want %v", s3Err.Message(), tt.wantMsg)
			}
		})
	}
}

func TestS3ErrorImplementsError(t *testing.T) {
	err := ErrNoSuchBucket
	if err.Error() == "" {
		t.Error("Error() should return non-empty string")
	}
}

func TestS3ErrorWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewS3Error(ErrInternal, cause)

	if !errors.Is(err, ErrInternal) {
		t.Error("errors.Is should return true for ErrInternal")
	}

	if !errors.Is(err, cause) {
		t.Error("errors.Is should return true for underlying cause")
	}
}

func TestNewS3Error(t *testing.T) {
	err := NewS3Error(ErrNoSuchBucket)
	if !errors.Is(err, ErrNoSuchBucket) {
		t.Error("NewS3Error should wrap the error")
	}

	s3Err, ok := err.(S3Error)
	if !ok {
		t.Fatal("error should be S3Error")
	}

	if s3Err.Code() != "NoSuchBucket" {
		t.Errorf("Code() = %v, want NoSuchBucket", s3Err.Code())
	}
}
