package engine

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/openendpoint/openendpoint/internal/metadata"
	"github.com/openendpoint/openendpoint/internal/metadata/pebble"
	"github.com/openendpoint/openendpoint/internal/storage/flatfile"
)

func setupTestEngine(t *testing.T) (*ObjectService, func()) {
	dir, err := os.MkdirTemp("", "openendpoint-engine-test-*")
	if err != nil {
		t.Fatal(err)
	}

	storage, err := flatfile.New(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}

	metadata, err := pebble.New(dir)
	if err != nil {
		storage.Close()
		os.RemoveAll(dir)
		t.Fatal(err)
	}

	engine := New(storage, metadata, nil)

	cleanup := func() {
		engine.Close()
		os.RemoveAll(dir)
	}

	return engine, cleanup
}

func TestObjectServicePutGet(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a bucket
	err := engine.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}

	// Put an object
	content := []byte("Hello, World!")
	err = engine.PutObject(ctx, "test-bucket", "test-key", content, &PutObjectOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	// Get the object
	data, err := engine.GetObject(ctx, "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("GetObject() error = %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("GetObject() = %v, want %v", string(data), string(content))
	}
}

func TestObjectServiceDelete(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a bucket
	err := engine.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}

	// Put an object
	err = engine.PutObject(ctx, "test-bucket", "test-key", []byte("test"), nil)
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	// Delete the object
	err = engine.DeleteObject(ctx, "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("DeleteObject() error = %v", err)
	}

	// Verify it's gone
	_, err = engine.GetObject(ctx, "test-bucket", "test-key")
	if err == nil {
		t.Error("GetObject should fail after deletion")
	}
}

func TestObjectServiceListBuckets(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create buckets
	bucketNames := []string{"bucket1", "bucket2", "bucket3"}
	for _, name := range bucketNames {
		err := engine.CreateBucket(ctx, name)
		if err != nil {
			t.Fatalf("CreateBucket(%s) error = %v", name, err)
		}
	}

	// List buckets
	buckets, err := engine.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("ListBuckets() error = %v", err)
	}

	if len(buckets) != len(bucketNames) {
		t.Errorf("ListBuckets() count = %d, want %d", len(buckets), len(bucketNames))
	}
}

func TestObjectServiceListObjects(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a bucket
	err := engine.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}

	// Put objects with different prefixes
	objects := []string{"dir1/file1.txt", "dir1/file2.txt", "dir2/file3.txt", "root.txt"}
	for _, key := range objects {
		err = engine.PutObject(ctx, "test-bucket", key, []byte("content"), nil)
		if err != nil {
			t.Fatalf("PutObject(%s) error = %v", key, err)
		}
	}

	// List objects
	result, err := engine.ListObjects(ctx, "test-bucket", ListObjectsOptions{
		Prefix: "dir1/",
	})
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}

	if len(result.Objects) != 2 {
		t.Errorf("ListObjects() count = %d, want 2", len(result.Objects))
	}
}

func TestObjectServiceBucketExists(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a bucket
	err := engine.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}

	// Check it exists
	exists, err := engine.BucketExists(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("BucketExists() error = %v", err)
	}
	if !exists {
		t.Error("BucketExists() should return true")
	}

	// Check non-existent bucket
	exists, err = engine.BucketExists(ctx, "non-existent")
	if err != nil {
		t.Fatalf("BucketExists() error = %v", err)
	}
	if exists {
		t.Error("BucketExists() should return false for non-existent bucket")
	}
}

func TestObjectServiceHeadObject(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a bucket
	err := engine.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}

	// Put an object
	content := []byte("Hello, World!")
	err = engine.PutObject(ctx, "test-bucket", "test-key", content, &PutObjectOptions{
		ContentType: "text/plain",
		ContentLength: int64(len(content)),
	})
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	// Head the object
	meta, err := engine.HeadObject(ctx, "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("HeadObject() error = %v", err)
	}

	if meta.Size != int64(len(content)) {
		t.Errorf("HeadObject().Size = %d, want %d", meta.Size, len(content))
	}
}

func TestObjectServiceDeleteBucket(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a bucket
	err := engine.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}

	// Delete the bucket
	err = engine.DeleteBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("DeleteBucket() error = %v", err)
	}

	// Verify it's gone
	exists, err := engine.BucketExists(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("BucketExists() error = %v", err)
	}
	if exists {
		t.Error("Bucket should not exist after deletion")
	}
}

func TestObjectServiceGetBucket(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a bucket
	err := engine.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}

	// Get bucket info
	bucket, err := engine.GetBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("GetBucket() error = %v", err)
	}

	if bucket.Name != "test-bucket" {
		t.Errorf("GetBucket().Name = %s, want test-bucket", bucket.Name)
	}
}

func TestObjectServiceMultiparts(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a bucket
	err := engine.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}

	// Initiate multipart upload
	uploadID, err := engine.InitiateMultipartUpload(ctx, "test-bucket", "test-key", nil)
	if err != nil {
		t.Fatalf("InitiateMultipartUpload() error = %v", err)
	}

	if uploadID == "" {
		t.Error("UploadID should not be empty")
	}

	// Upload parts
	part1Data := []byte("part1")
	err = engine.PutPart(ctx, "test-bucket", "test-key", uploadID, 1, part1Data)
	if err != nil {
		t.Fatalf("PutPart() error = %v", err)
	}

	part2Data := []byte("part2")
	err = engine.PutPart(ctx, "test-bucket", "test-key", uploadID, 2, part2Data)
	if err != nil {
		t.Fatalf("PutPart() error = %v", err)
	}

	// Complete multipart upload
	err = engine.CompleteMultipartUpload(ctx, "test-bucket", "test-key", uploadID, []metadata.PartInfo{
		{PartNumber: 1, ETag: "etag1"},
		{PartNumber: 2, ETag: "etag2"},
	})
	if err != nil {
		t.Fatalf("CompleteMultipartUpload() error = %v", err)
	}

	// Verify object
	data, err := engine.GetObject(ctx, "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("GetObject() error = %v", err)
	}

	expected := string(part1Data) + string(part2Data)
	if string(data) != expected {
		t.Errorf("GetObject() = %v, want %v", string(data), expected)
	}
}

func TestObjectServiceListMultipartUploads(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a bucket
	err := engine.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}

	// Initiate multipart uploads
	_, err = engine.InitiateMultipartUpload(ctx, "test-bucket", "key1", nil)
	if err != nil {
		t.Fatalf("InitiateMultipartUpload() error = %v", err)
	}

	_, err = engine.InitiateMultipartUpload(ctx, "test-bucket", "key2", nil)
	if err != nil {
		t.Fatalf("InitiateMultipartUpload() error = %v", err)
	}

	// List uploads
	result, err := engine.ListMultipartUpload(ctx, "test-bucket", "")
	if err != nil {
		t.Fatalf("ListMultipartUpload() error = %v", err)
	}

	if len(result.Uploads) != 2 {
		t.Errorf("ListMultipartUpload() count = %d, want 2", len(result.Uploads))
	}
}

func TestObjectServiceAbortMultipartUpload(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a bucket
	err := engine.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}

	// Initiate multipart upload
	uploadID, err := engine.InitiateMultipartUpload(ctx, "test-bucket", "test-key", nil)
	if err != nil {
		t.Fatalf("InitiateMultipartUpload() error = %v", err)
	}

	// Abort the upload
	err = engine.AbortMultipartUpload(ctx, "test-bucket", "test-key", uploadID)
	if err != nil {
		t.Fatalf("AbortMultipartUpload() error = %v", err)
	}

	// List uploads - should be empty
	result, err := engine.ListMultipartUpload(ctx, "test-bucket", "")
	if err != nil {
		t.Fatalf("ListMultipartUpload() error = %v", err)
	}

	if len(result.Uploads) != 0 {
		t.Errorf("ListMultipartUpload() count = %d, want 0", len(result.Uploads))
	}
}

func TestValidateBucketName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid-bucket", false},
		{"my-bucket-name", false},
		{"ab", true},               // too short
		{"bucket", false},          // minimum 3 chars
		{"a" + strings.Repeat("a", 64), true}, // too long (>63)
		{"192.168.1.1", false},     // IP-style name is valid
		{"-invalid", true},         // starts with hyphen
		{"invalid-", true},         // ends with hyphen
		{"invalid..name", true},    // consecutive dots
		{"invalid.-name", true},    // dot followed by hyphen
		{"-invalid-name", true},    // starts with hyphen
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBucketName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBucketName(%s) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestValidateObjectKey(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid/key", false},
		{"my-object-name", false},
		{"key/with/multiple/parts", false},
		{"", true},                   // empty key
		{"/starts/with/slash", false}, // leading slash is OK (handled by path)
		{"key/with/dot..dot", false},  // dots are OK
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateObjectKey(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateObjectKey(%s) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
