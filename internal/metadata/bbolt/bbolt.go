package bbolt

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/openendpoint/openendpoint/internal/metadata"
	bolt "go.etcd.io/bbolt"
)

// BBoltStore implements metadata.Store using bbolt
type BBoltStore struct {
	db *bolt.DB
}

// New creates a new bbolt metadata store
func New(rootDir string) (*BBoltStore, error) {
	dbPath := filepath.Join(rootDir, "metadata.db")

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bbolt database: %w", err)
	}

	// Create buckets
	err = db.Update(func(tx *bolt.Tx) error {
		// Buckets bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("buckets")); err != nil {
			return err
		}
		// Objects bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("objects")); err != nil {
			return err
		}
		// Multipart uploads bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("multipart")); err != nil {
			return err
		}
		// Parts bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("parts")); err != nil {
			return err
		}
		// Lifecycle bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("lifecycle")); err != nil {
			return err
		}
		// Versioning bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("versioning")); err != nil {
			return err
		}
		// Replication bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("replication")); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create buckets: %w", err)
	}

	return &BBoltStore{db: db}, nil
}

// CreateBucket creates a new bucket
func (b *BBoltStore) CreateBucket(ctx context.Context, bucket string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		buckets := tx.Bucket([]byte("buckets"))
		meta := &metadata.BucketMetadata{
			Name:         bucket,
			CreationDate: nowUnix(),
		}
		return buckets.Put([]byte(bucket), mustEncode(meta))
	})
}

// DeleteBucket deletes a bucket
func (b *BBoltStore) DeleteBucket(ctx context.Context, bucket string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		buckets := tx.Bucket([]byte("buckets"))
		return buckets.Delete([]byte(bucket))
	})
}

// GetBucket gets bucket metadata
func (b *BBoltStore) GetBucket(ctx context.Context, bucket string) (*metadata.BucketMetadata, error) {
	var meta metadata.BucketMetadata
	err := b.db.View(func(tx *bolt.Tx) error {
		buckets := tx.Bucket([]byte("buckets"))
		data := buckets.Get([]byte(bucket))
		if data == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}
		return mustDecode(data, &meta)
	})
	return &meta, err
}

// ListBuckets lists all buckets
func (b *BBoltStore) ListBuckets(ctx context.Context) ([]string, error) {
	var buckets []string
	err := b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("buckets"))
		return bkt.ForEach(func(k, v []byte) error {
			buckets = append(buckets, string(k))
			return nil
		})
	})
	return buckets, err
}

// PutObject stores object metadata
func (b *BBoltStore) PutObject(ctx context.Context, bucket, key string, meta *metadata.ObjectMetadata) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		objects := tx.Bucket([]byte("objects"))
		objKey := bucket + "/" + key
		return objects.Put([]byte(objKey), mustEncode(meta))
	})
}

// GetObject gets object metadata
func (b *BBoltStore) GetObject(ctx context.Context, bucket, key string, versionID string) (*metadata.ObjectMetadata, error) {
	var meta metadata.ObjectMetadata
	err := b.db.View(func(tx *bolt.Tx) error {
		objects := tx.Bucket([]byte("objects"))
		objKey := bucket + "/" + key
		data := objects.Get([]byte(objKey))
		if data == nil {
			return fmt.Errorf("object not found: %s/%s", bucket, key)
		}
		return mustDecode(data, &meta)
	})
	return &meta, err
}

// DeleteObject deletes object metadata
func (b *BBoltStore) DeleteObject(ctx context.Context, bucket, key string, versionID string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		objects := tx.Bucket([]byte("objects"))
		objKey := bucket + "/" + key
		return objects.Delete([]byte(objKey))
	})
}

// ListObjects lists objects with optional prefix
func (b *BBoltStore) ListObjects(ctx context.Context, bucket, prefix string, opts metadata.ListOptions) ([]metadata.ObjectMetadata, error) {
	var objects []metadata.ObjectMetadata
	err := b.db.View(func(tx *bolt.Tx) error {
		objectsBkt := tx.Bucket([]byte("objects"))
		prefixKey := bucket + "/" + prefix

		maxKeys := opts.MaxKeys
		if maxKeys == 0 {
			maxKeys = 1000
		}

		cursor := objectsBkt.Cursor()
		for k, v := cursor.Seek([]byte(prefixKey)); k != nil && len(objects) < maxKeys; k, v = cursor.Next() {
			key := string(k)
			if len(key) < len(bucket)+1 || key[:len(bucket)+1] != bucket+"/" {
				break
			}

			var meta metadata.ObjectMetadata
			if err := mustDecode(v, &meta); err != nil {
				continue
			}
			objects = append(objects, meta)
		}
		return nil
	})
	return objects, err
}

// CreateMultipartUpload creates a new multipart upload
func (b *BBoltStore) CreateMultipartUpload(ctx context.Context, bucket, key, uploadID string, meta *metadata.ObjectMetadata) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		multipart := tx.Bucket([]byte("multipart"))
		multiMeta := &metadata.MultipartUploadMetadata{
			UploadID:  uploadID,
			Key:       key,
			Bucket:    bucket,
			Initiated: nowUnix(),
			Metadata:  meta.Metadata,
		}
		multiKey := bucket + "/" + key + "/" + uploadID
		return multipart.Put([]byte(multiKey), mustEncode(multiMeta))
	})
}

// PutPart stores part metadata
func (b *BBoltStore) PutPart(ctx context.Context, bucket, key, uploadID string, partNumber int, partMeta *metadata.PartMetadata) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		parts := tx.Bucket([]byte("parts"))
		partKey := fmt.Sprintf("%s/%s/%s/%d", bucket, key, uploadID, partNumber)
		return parts.Put([]byte(partKey), mustEncode(partMeta))
	})
}

// CompleteMultipartUpload completes a multipart upload
func (b *BBoltStore) CompleteMultipartUpload(ctx context.Context, bucket, key, uploadID string, parts []metadata.PartInfo) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		multipart := tx.Bucket([]byte("multipart"))
		partsBkt := tx.Bucket([]byte("parts"))

		multiKey := bucket + "/" + key + "/" + uploadID
		if err := multipart.Delete([]byte(multiKey)); err != nil {
			return err
		}

		// Delete parts
		for i := 1; i <= len(parts); i++ {
			partKey := fmt.Sprintf("%s/%s/%s/%d", bucket, key, uploadID, i)
			partsBkt.Delete([]byte(partKey))
		}
		return nil
	})
}

// AbortMultipartUpload aborts a multipart upload
func (b *BBoltStore) AbortMultipartUpload(ctx context.Context, bucket, key, uploadID string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		multipart := tx.Bucket([]byte("multipart"))
		multiKey := bucket + "/" + key + "/" + uploadID
		return multipart.Delete([]byte(multiKey))
	})
}

// ListParts lists parts of a multipart upload
func (b *BBoltStore) ListParts(ctx context.Context, bucket, key, uploadID string) ([]metadata.PartMetadata, error) {
	var parts []metadata.PartMetadata
	err := b.db.View(func(tx *bolt.Tx) error {
		partsBkt := tx.Bucket([]byte("parts"))
		prefix := fmt.Sprintf("%s/%s/%s/", bucket, key, uploadID)

		cursor := partsBkt.Cursor()
		for k, v := cursor.Seek([]byte(prefix)); k != nil; k, v = cursor.Next() {
			key := string(k)
			if len(key) < len(prefix) || key[:len(prefix)] != prefix {
				break
			}

			var partMeta metadata.PartMetadata
			if err := mustDecode(v, &partMeta); err != nil {
				continue
			}
			parts = append(parts, partMeta)
		}
		return nil
	})
	return parts, err
}

// ListMultipartUploads lists multipart uploads
func (b *BBoltStore) ListMultipartUploads(ctx context.Context, bucket, prefix string) ([]metadata.MultipartUploadMetadata, error) {
	var uploads []metadata.MultipartUploadMetadata
	err := b.db.View(func(tx *bolt.Tx) error {
		multipart := tx.Bucket([]byte("multipart"))
		if multipart == nil {
			return nil
		}
		prefixKey := bucket + "/" + prefix

		cursor := multipart.Cursor()
		for k, v := cursor.Seek([]byte(prefixKey)); k != nil; k, v = cursor.Next() {
			key := string(k)
			// Check if still within the bucket prefix
			if !containsPrefix(key, bucket+"/") {
				break
			}

			var meta metadata.MultipartUploadMetadata
			if err := mustDecode(v, &meta); err != nil {
				continue
			}
			uploads = append(uploads, meta)
		}
		return nil
	})
	return uploads, err
}

// containsPrefix checks if string contains the given prefix
func containsPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// PutLifecycleRule puts a lifecycle rule
func (b *BBoltStore) PutLifecycleRule(ctx context.Context, bucket string, rule *metadata.LifecycleRule) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		lifecycle := tx.Bucket([]byte("lifecycle"))
		return lifecycle.Put([]byte(bucket+"_"+rule.ID), mustEncode(rule))
	})
}

// GetLifecycleRules gets lifecycle rules for a bucket
func (b *BBoltStore) GetLifecycleRules(ctx context.Context, bucket string) ([]metadata.LifecycleRule, error) {
	var rules []metadata.LifecycleRule
	err := b.db.View(func(tx *bolt.Tx) error {
		lifecycle := tx.Bucket([]byte("lifecycle"))
		prefix := []byte(bucket + "_")

		cursor := lifecycle.Cursor()
		for k, v := cursor.Seek(prefix); k != nil; k, v = cursor.Next() {
			key := string(k)
			if len(key) < len(prefix) || key[:len(prefix)] != bucket+"_" {
				break
			}

			var rule metadata.LifecycleRule
			if err := mustDecode(v, &rule); err != nil {
				continue
			}
			rules = append(rules, rule)
		}
		return nil
	})
	return rules, err
}

// PutBucketVersioning puts bucket versioning configuration
func (b *BBoltStore) PutBucketVersioning(ctx context.Context, bucket string, versioning *metadata.BucketVersioning) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		versioningBkt := tx.Bucket([]byte("versioning"))
		return versioningBkt.Put([]byte(bucket), mustEncode(versioning))
	})
}

// GetBucketVersioning gets bucket versioning configuration
func (b *BBoltStore) GetBucketVersioning(ctx context.Context, bucket string) (*metadata.BucketVersioning, error) {
	var versioning metadata.BucketVersioning
	err := b.db.View(func(tx *bolt.Tx) error {
		versioningBkt := tx.Bucket([]byte("versioning"))
		data := versioningBkt.Get([]byte(bucket))
		if data == nil {
			return nil
		}
		return mustDecode(data, &versioning)
	})
	return &versioning, err
}

// PutReplicationConfig stores replication configuration
func (b *BBoltStore) PutReplicationConfig(ctx context.Context, bucket string, config *metadata.ReplicationConfig) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		replicationBkt := tx.Bucket([]byte("replication"))
		return replicationBkt.Put([]byte(bucket), mustEncode(config))
	})
}

// GetReplicationConfig gets replication configuration
func (b *BBoltStore) GetReplicationConfig(ctx context.Context, bucket string) (*metadata.ReplicationConfig, error) {
	var config metadata.ReplicationConfig
	err := b.db.View(func(tx *bolt.Tx) error {
		replicationBkt := tx.Bucket([]byte("replication"))
		data := replicationBkt.Get([]byte(bucket))
		if data == nil {
			return nil
		}
		return mustDecode(data, &config)
	})
	return &config, err
}

// DeleteReplicationConfig deletes replication configuration
func (b *BBoltStore) DeleteReplicationConfig(ctx context.Context, bucket string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		replicationBkt := tx.Bucket([]byte("replication"))
		return replicationBkt.Delete([]byte(bucket))
	})
}

// Close closes the store
func (b *BBoltStore) Close() error {
	return b.db.Close()
}

// nowUnix returns current Unix timestamp
func nowUnix() int64 {
	return time.Now().Unix()
}

// mustEncode panics on encode error
func mustEncode(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// mustDecode panics on decode error
func mustDecode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
