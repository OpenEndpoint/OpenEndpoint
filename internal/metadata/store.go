package metadata

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"time"
)

// CORSConfiguration represents S3 CORS configuration
type CORSConfiguration struct {
	XMLName   xml.Name    `xml:"CORSConfiguration"`
	CORSRules []CORSRule `xml:"CORSRule"`
}

// CORSRule represents a single CORS rule
type CORSRule struct {
	AllowedMethods []string `xml:"AllowedMethod"`
	AllowedOrigins []string `xml:"AllowedOrigin"`
	AllowedHeaders []string `xml:"AllowedHeader,omitempty"`
	ExposeHeaders  []string `xml:"ExposeHeader,omitempty"`
	MaxAgeSeconds  int      `xml:"MaxAgeSeconds,omitempty"`
}

// Store defines the interface for metadata storage
type Store interface {
	// Bucket operations
	CreateBucket(ctx context.Context, bucket string) error
	DeleteBucket(ctx context.Context, bucket string) error
	GetBucket(ctx context.Context, bucket string) (*BucketMetadata, error)
	ListBuckets(ctx context.Context) ([]string, error)

	// Object operations
	PutObject(ctx context.Context, bucket, key string, meta *ObjectMetadata) error
	GetObject(ctx context.Context, bucket, key string, versionID string) (*ObjectMetadata, error)
	DeleteObject(ctx context.Context, bucket, key string, versionID string) error
	ListObjects(ctx context.Context, bucket, prefix string, opts ListOptions) ([]ObjectMetadata, error)

	// Multipart upload operations
	CreateMultipartUpload(ctx context.Context, bucket, key, uploadID string, meta *ObjectMetadata) error
	PutPart(ctx context.Context, bucket, key, uploadID string, partNumber int, meta *PartMetadata) error
	CompleteMultipartUpload(ctx context.Context, bucket, key, uploadID string, parts []PartInfo) error
	AbortMultipartUpload(ctx context.Context, bucket, key, uploadID string) error
	ListParts(ctx context.Context, bucket, key, uploadID string) ([]PartMetadata, error)
	ListMultipartUploads(ctx context.Context, bucket, prefix string) ([]MultipartUploadMetadata, error)

	// Lifecycle operations
	PutLifecycleRule(ctx context.Context, bucket string, rule *LifecycleRule) error
	GetLifecycleRules(ctx context.Context, bucket string) ([]LifecycleRule, error)
	DeleteLifecycleRule(ctx context.Context, bucket, ruleID string) error

	// Replication operations
	PutReplicationConfig(ctx context.Context, bucket string, config *ReplicationConfig) error
	GetReplicationConfig(ctx context.Context, bucket string) (*ReplicationConfig, error)
	DeleteReplicationConfig(ctx context.Context, bucket string) error

	// Versioning operations
	PutBucketVersioning(ctx context.Context, bucket string, versioning *BucketVersioning) error
	GetBucketVersioning(ctx context.Context, bucket string) (*BucketVersioning, error)

	// CORS operations
	PutBucketCors(ctx context.Context, bucket string, cors *CORSConfiguration) error
	GetBucketCors(ctx context.Context, bucket string) (*CORSConfiguration, error)
	DeleteBucketCors(ctx context.Context, bucket string) error

	// Policy operations
	PutBucketPolicy(ctx context.Context, bucket string, policy *string) error
	GetBucketPolicy(ctx context.Context, bucket string) (*string, error)
	DeleteBucketPolicy(ctx context.Context, bucket string) error

	// Encryption operations
	PutBucketEncryption(ctx context.Context, bucket string, encryption *BucketEncryption) error
	GetBucketEncryption(ctx context.Context, bucket string) (*BucketEncryption, error)
	DeleteBucketEncryption(ctx context.Context, bucket string) error

	// Tagging operations
	PutBucketTags(ctx context.Context, bucket string, tags map[string]string) error
	GetBucketTags(ctx context.Context, bucket string) (map[string]string, error)
	DeleteBucketTags(ctx context.Context, bucket string) error

	// Object Lock operations
	PutObjectLock(ctx context.Context, bucket string, config *ObjectLockConfig) error
	GetObjectLock(ctx context.Context, bucket string) (*ObjectLockConfig, error)
	DeleteObjectLock(ctx context.Context, bucket string) error

	// Object Retention operations
	PutObjectRetention(ctx context.Context, bucket, key string, retention *ObjectRetention) error
	GetObjectRetention(ctx context.Context, bucket, key string) (*ObjectRetention, error)

	// Object Legal Hold operations
	PutObjectLegalHold(ctx context.Context, bucket, key string, legalHold *ObjectLegalHold) error
	GetObjectLegalHold(ctx context.Context, bucket, key string) (*ObjectLegalHold, error)

	// PublicAccessBlock operations
	PutPublicAccessBlock(ctx context.Context, bucket string, config *PublicAccessBlockConfiguration) error
	GetPublicAccessBlock(ctx context.Context, bucket string) (*PublicAccessBlockConfiguration, error)
	DeletePublicAccessBlock(ctx context.Context, bucket string) error

	// Accelerate operations
	PutBucketAccelerate(ctx context.Context, bucket string, config *BucketAccelerateConfiguration) error
	GetBucketAccelerate(ctx context.Context, bucket string) (*BucketAccelerateConfiguration, error)
	DeleteBucketAccelerate(ctx context.Context, bucket string) error

	// Inventory operations
	PutBucketInventory(ctx context.Context, bucket, id string, config *InventoryConfiguration) error
	GetBucketInventory(ctx context.Context, bucket, id string) (*InventoryConfiguration, error)
	ListBucketInventory(ctx context.Context, bucket string) ([]InventoryConfiguration, error)
	DeleteBucketInventory(ctx context.Context, bucket, id string) error

	// Analytics operations
	PutBucketAnalytics(ctx context.Context, bucket, id string, config *AnalyticsConfiguration) error
	GetBucketAnalytics(ctx context.Context, bucket, id string) (*AnalyticsConfiguration, error)
	ListBucketAnalytics(ctx context.Context, bucket string) ([]AnalyticsConfiguration, error)
	DeleteBucketAnalytics(ctx context.Context, bucket, id string) error

	// Presigned URL operations
	PutPresignedURL(ctx context.Context, url string, req *PresignedURLRequest) error
	GetPresignedURL(ctx context.Context, url string) (*PresignedURLRequest, error)
	DeletePresignedURL(ctx context.Context, url string) error

	// Website operations
	PutBucketWebsite(ctx context.Context, bucket string, config *WebsiteConfiguration) error
	GetBucketWebsite(ctx context.Context, bucket string) (*WebsiteConfiguration, error)
	DeleteBucketWebsite(ctx context.Context, bucket string) error

	// Notification operations
	PutBucketNotification(ctx context.Context, bucket string, config *NotificationConfiguration) error
	GetBucketNotification(ctx context.Context, bucket string) (*NotificationConfiguration, error)
	DeleteBucketNotification(ctx context.Context, bucket string) error

	// Logging operations
	PutBucketLogging(ctx context.Context, bucket string, config *LoggingConfiguration) error
	GetBucketLogging(ctx context.Context, bucket string) (*LoggingConfiguration, error)
	DeleteBucketLogging(ctx context.Context, bucket string) error

	// Location operations
	PutBucketLocation(ctx context.Context, bucket string, location string) error
	GetBucketLocation(ctx context.Context, bucket string) (string, error)

	// Ownership controls operations
	PutBucketOwnershipControls(ctx context.Context, bucket string, config *OwnershipControls) error
	GetBucketOwnershipControls(ctx context.Context, bucket string) (*OwnershipControls, error)
	DeleteBucketOwnershipControls(ctx context.Context, bucket string) error

	// Metrics operations
	PutBucketMetrics(ctx context.Context, bucket string, id string, config *MetricsConfiguration) error
	GetBucketMetrics(ctx context.Context, bucket string, id string) (*MetricsConfiguration, error)
	DeleteBucketMetrics(ctx context.Context, bucket string, id string) error
	ListBucketMetrics(ctx context.Context, bucket string) ([]MetricsConfiguration, error)

	// Close closes the store
	Close() error
}

// BucketMetadata contains bucket-level metadata
type BucketMetadata struct {
	Name          string    `json:"name"`
	CreationDate int64     `json:"creation_date"`
	Owner         string    `json:"owner"`
	Region        string    `json:"region"`
}

// ObjectMetadata contains object-level metadata
type ObjectMetadata struct {
	Key             string            `json:"key"`
	Bucket          string            `json:"bucket"`
	Size            int64             `json:"size"`
	ETag            string            `json:"etag"`
	ContentType     string            `json:"content_type"`
	ContentEncoding string            `json:"content_encoding"`
	CacheControl    string            `json:"cache_control"`
	Metadata        map[string]string `json:"metadata"`
	StorageClass    string            `json:"storage_class"`
	VersionID       string            `json:"version_id"`
	IsLatest        bool              `json:"is_latest"`
	IsDeleteMarker  bool              `json:"is_delete_marker"`
	LastModified    int64             `json:"last_modified"`
	Expires         int64             `json:"expires"`
	Parts           []PartInfo        `json:"parts,omitempty"`
}

// PartInfo represents a part in a multipart upload
type PartInfo struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
}

// PartMetadata contains metadata for a part
type PartMetadata struct {
	UploadID     string `json:"upload_id"`
	Key          string `json:"key"`
	Bucket       string `json:"bucket"`
	PartNumber   int    `json:"part_number"`
	ETag         string `json:"etag"`
	Size         int64  `json:"size"`
	LastModified int64  `json:"last_modified"`
}

// MultipartUploadMetadata contains metadata for a multipart upload
type MultipartUploadMetadata struct {
	UploadID  string            `json:"upload_id"`
	Key      string            `json:"key"`
	Bucket   string            `json:"bucket"`
	Initiated int64            `json:"initiated"`
	Metadata map[string]string `json:"metadata"`
}

// LifecycleRule defines a lifecycle rule
type LifecycleRule struct {
	ID         string     `json:"id"`
	Prefix     string     `json:"prefix"`
	Status     string     `json:"status"` // Enabled or Disabled
	Expiration *Expiration `json:"expiration,omitempty"`
	Transitions []Transition `json:"transitions,omitempty"`
	NoncurrentVersionExpiration *NoncurrentVersionExpiration `json:"noncurrent_version_expiration,omitempty"`
}

type Expiration struct {
	Days          int  `json:"days"`
	Date          int64 `json:"date"`
	ExpiredObjectDeleteMarker bool `json:"expired_object_delete_marker"`
}

type Transition struct {
	Days          int    `json:"days"`
	StorageClass string `json:"storage_class"`
	Date         int64  `json:"date"`
}

type NoncurrentVersionExpiration struct {
	NoncurrentDays int `json:"noncurrent_days"`
}

// BucketEncryption contains bucket encryption configuration
type BucketEncryption struct {
	Rule        EncryptionRule `json:"Rule"`
}

// EncryptionRule contains encryption rule
type EncryptionRule struct {
	Apply       ApplyEncryptionConfiguration `json:"Apply"`
}

// ApplyEncryptionConfiguration applies encryption configuration
type ApplyEncryptionConfiguration struct {
	SSEAlgorithm         string `json:"SSEAlgorithm,omitempty"`
	KMSMasterKeyID      string `json:"KMSMasterKeyID,omitempty"`
}

// ObjectLockConfig contains object lock configuration
type ObjectLockConfig struct {
	Enabled bool `json:"Enabled"`
}

// ObjectRetention contains object retention configuration
type ObjectRetention struct {
	Mode     string `json:"Mode"`     // GOVERNANCE, COMPLIANCE
	RetainUntilDate int64 `json:"RetainUntilDate"`
}

// ObjectLegalHold contains object legal hold configuration
type ObjectLegalHold struct {
	Status string `json:"Status"` // ON, OFF
}

// PublicAccessBlockConfiguration contains public access block configuration
type PublicAccessBlockConfiguration struct {
	BlockPublicAcls       bool `json:"BlockPublicAcls"`
	BlockPublicPolicy     bool `json:"BlockPublicPolicy"`
	IgnorePublicAcls      bool `json:"IgnorePublicAcls"`
	RestrictPublicBuckets bool `json:"RestrictPublicBuckets"`
}

// BucketAccelerateConfiguration contains bucket accelerate configuration
type BucketAccelerateConfiguration struct {
	Status string `json:"Status"` // Enabled or Suspended
}

// InventoryConfiguration contains bucket inventory configuration
type InventoryConfiguration struct {
	ID        string          `json:"Id"`
	Enabled   bool            `json:"Enabled"`
	Filter    InventoryFilter `json:"Filter,omitempty"`
	Destination InventoryDestination `json:"Destination"`
	Schedule  InventorySchedule `json:"Schedule"`
	IncludedFields []string `json:"IncludedFields,omitempty"`
	OptionalFields []string `json:"OptionalFields,omitempty"`
}

type InventoryFilter struct {
	Prefix string `json:"Prefix,omitempty"`
}

type InventoryDestination struct {
	Bucket BucketDestination `json:"Bucket"`
}

type BucketDestination struct {
	Format string `json:"Format"` // CSV, ORC, Parquet
	Prefix string `json:"Prefix,omitempty"`
	Account string `json:"Account,omitempty"`
	Arn    string `json:"Arn"`
}

type InventorySchedule struct {
	Frequency string `json:"Frequency"` // Daily, Weekly
}

// AnalyticsConfiguration contains bucket analytics configuration
type AnalyticsConfiguration struct {
	ID        string              `json:"Id"`
	Enabled   bool                `json:"Enabled"`
	Filter    AnalyticsFilter     `json:"Filter,omitempty"`
	StorageClassAnalysis AnalyticsStorageClassAnalysis `json:"StorageClassAnalysis"`
}

type AnalyticsFilter struct {
	Prefix string `json:"Prefix,omitempty"`
}

type AnalyticsStorageClassAnalysis struct {
	DataExport AnalyticsDataExport `json:"DataExport"`
}

type AnalyticsDataExport struct {
	OutputSchemaVersion string            `json:"OutputSchemaVersion"`
	Destination        AnalyticsDestination `json:"Destination"`
}

type AnalyticsDestination struct {
	BucketDestination AnalyticsBucketDestination `json:"Bucket"`
}

type AnalyticsBucketDestination struct {
	Format     string `json:"Format"` // CSV
	Prefix     string `json:"Prefix,omitempty"`
	Account    string `json:"Account,omitempty"`
	Arn        string `json:"Arn"`
}

// WebsiteConfiguration contains bucket website configuration
type WebsiteConfiguration struct {
	IndexDocument *IndexDocument `json:"IndexDocument,omitempty"`
	ErrorDocument *ErrorDocument `json:"ErrorDocument,omitempty"`
	RoutingRules  []RoutingRule `json:"RoutingRules,omitempty"`
}

// IndexDocument specifies the default index page
type IndexDocument struct {
	Suffix string `json:"Suffix"`
}

// ErrorDocument specifies the error page
type ErrorDocument struct {
	Key string `json:"Key"`
}

// RoutingRule represents a single routing rule
type RoutingRule struct {
	Condition *RoutingCondition `json:"Condition,omitempty"`
	Redirect  *RoutingRedirect  `json:"Redirect,omitempty"`
}

// RoutingCondition specifies when a routing rule is applied
type RoutingCondition struct {
	KeyPrefixEquals              string `json:"KeyPrefixEquals,omitempty"`
	HttpErrorCodeReturnedEquals string `json:"HttpErrorCodeReturnedEquals,omitempty"`
}

// RoutingRedirect specifies how to redirect
type RoutingRedirect struct {
	Protocol           string `json:"Protocol,omitempty"`
	HostName           string `json:"HostName,omitempty"`
	ReplaceKeyPrefixWith string `json:"ReplaceKeyPrefixWith,omitempty"`
	ReplaceKeyWith     string `json:"ReplaceKeyWith,omitempty"`
	HttpRedirectCode   string `json:"HttpRedirectCode,omitempty"`
}

// NotificationConfiguration contains bucket notification configuration
type NotificationConfiguration struct {
	TopicConfigurations []TopicConfiguration `json:"TopicConfigurations,omitempty"`
	QueueConfigurations []QueueConfiguration `json:"QueueConfigurations,omitempty"`
	LambdaFunctionConfigurations []LambdaFunctionConfiguration `json:"LambdaFunctionConfigurations,omitempty"`
}

// TopicConfiguration contains SNS topic notification configuration
type TopicConfiguration struct {
	ID        string   `json:"Id"`
	Topic     string   `json:"Topic"`
	Events    []string `json:"Event"`
	FilterRules []FilterRule `json:"FilterRules,omitempty"`
}

// QueueConfiguration contains SQS queue notification configuration
type QueueConfiguration struct {
	ID        string   `json:"Id"`
	Queue     string   `json:"Queue"`
	Events    []string `json:"Event"`
	FilterRules []FilterRule `json:"FilterRules,omitempty"`
}

// LambdaFunctionConfiguration contains Lambda notification configuration
type LambdaFunctionConfiguration struct {
	ID           string   `json:"Id"`
	Function     string   `json:"Function"`
	Events       []string `json:"Event"`
	FilterRules  []FilterRule `json:"FilterRules,omitempty"`
}

// FilterRule contains notification filter rules
type FilterRule struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

// LoggingConfiguration contains bucket logging configuration
type LoggingConfiguration struct {
	LoggingEnabled  bool           `json:"LoggingEnabled"`
	TargetBucket   string         `json:"TargetBucket,omitempty"`
	TargetPrefix  string         `json:"TargetPrefix,omitempty"`
	TargetGrants  []AccessGrant `json:"TargetGrants,omitempty"`
}

// AccessGrant contains access control grant information
type AccessGrant struct {
	Grantee    *Grantee    `json:"Grantee,omitempty"`
	Permission string      `json:"Permission"`
}

// Grantee contains grantee information
type Grantee struct {
	Type          string `json:"Type"`
	DisplayName   string `json:"DisplayName,omitempty"`
	EmailAddress  string `json:"EmailAddress,omitempty"`
	ID            string `json:"ID,omitempty"`
	URI          string `json:"URI,omitempty"`
}

// BucketVersioning contains versioning configuration
type BucketVersioning struct {
	Status    string `json:"status"` // Enabled, Suspended, or ""
	MFADelete string `json:"mfa_delete"` // Enabled or Disabled
}

// OwnershipControls contains bucket ownership controls
type OwnershipControls struct {
	Rules []OwnershipRule `json:"Rules"`
}

// OwnershipRule contains an ownership rule
type OwnershipRule struct {
	ObjectOwnership string `json:"ObjectOwnership"` // ObjectWriter, BucketOwnerPreferred, BucketOwnerEnforced
}

// MetricsConfiguration contains bucket metrics configuration
type MetricsConfiguration struct {
	ID        string          `json:"Id"`
	Enabled   bool            `json:"Enabled"`
	Filter    *MetricsFilter  `json:"Filter,omitempty"`
}

// MetricsFilter contains metrics filter
type MetricsFilter struct {
	Prefix string            `json:"Prefix,omitempty"`
	Tag    map[string]string `json:"Tag,omitempty"`
}

// ReplicationConfig contains bucket replication configuration
type ReplicationConfig struct {
	Role    string              `json:"role"`
	Rules   []ReplicationRule  `json:"rules"`
}

// ReplicationRule contains a replication rule
type ReplicationRule struct {
	ID        string `json:"id"`
	Status    string `json:"status"` // Enabled or Disabled
	Prefix    string `json:"prefix"`
	Destination Destination `json:"destination"`
}

// Destination contains replication destination
type Destination struct {
	Bucket       string `json:"bucket"`
	StorageClass string `json:"storage_class,omitempty"`
}

// ListOptions contains options for listing objects
type ListOptions struct {
	Prefix       string
	Delimiter    string
	MaxKeys      int
	Marker       string
	VersionIDMarker string
}

// MarshalJSON implements custom JSON marshaling
func (o *ObjectMetadata) MarshalJSON() ([]byte, error) {
	type Alias ObjectMetadata
	return json.Marshal(&struct {
		*Alias
		LastModified time.Time `json:"last_modified,omitempty"`
	}{
		Alias:        (*Alias)(o),
		LastModified: time.Unix(o.LastModified, 0).UTC(),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling
func (o *ObjectMetadata) UnmarshalJSON(data []byte) error {
	type Alias ObjectMetadata
	aux := &struct {
		*Alias
		LastModified time.Time `json:"last_modified"`
	}{
		Alias: (*Alias)(o),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if !aux.LastModified.IsZero() {
		o.LastModified = aux.LastModified.Unix()
	}
	return nil
}

// PresignedURLRequest represents a request to generate a presigned URL
type PresignedURLRequest struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	Method    string `json:"method"` // GET, PUT, DELETE, etc.
	Expires   int64  `json:"expires"` // Expiration time in seconds from now
	Scheme    string `json:"scheme"` // http or https
	Host      string `json:"host"`
	Headers   map[string]string `json:"headers,omitempty"`
	ReponseHeaders map[string]string `json:"response_headers,omitempty"`
}

// PresignedURLResponse contains the generated presigned URL
type PresignedURLResponse struct {
	URL               string            `json:"url"`
	SignedHeaders     []string          `json:"signed_headers"`
	Expiration        string            `json:"expiration"`
	ExpiresAt         int64             `json:"expires_at"`
}
