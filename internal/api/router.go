package api

import (
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openendpoint/openendpoint/internal/auth"
	"github.com/openendpoint/openendpoint/internal/config"
	"github.com/openendpoint/openendpoint/internal/engine"
	"github.com/openendpoint/openendpoint/internal/storage"
	s3types "github.com/openendpoint/openendpoint/pkg/s3types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	s3RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openendpoint_s3_requests_total",
			Help: "Total S3 API requests",
		},
		[]string{"operation", "status"},
	)
)

type Router struct {
	engine      *engine.ObjectService
	auth        *auth.Auth
	logger      *zap.SugaredLogger
	config      *config.Config
}

// NewRouter creates a new S3 API router
func NewRouter(engine *engine.ObjectService, auth *auth.Auth, logger *zap.SugaredLogger, cfg *config.Config) *Router {
	return &Router{
		engine: engine,
		auth:   auth,
		logger: logger,
		config: cfg,
	}
}

// ServeHTTP handles S3 API requests
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, HEAD, POST")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	// Handle OPTIONS for CORS preflight
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Strip /s3/ prefix
	path := strings.TrimPrefix(req.URL.Path, "/s3/")
	if path == req.URL.Path {
		// No prefix found, try without
		path = req.URL.Path
	}

	// Parse bucket and key
	bucket, key, err := parseBucketKey(req, path)
	if err != nil {
		r.writeError(w, ErrInvalidURI)
		return
	}

	// Route request
	r.route(w, req, bucket, key)
}

func (r *Router) route(w http.ResponseWriter, req *http.Request, bucket, key string) {
	// Get operation from query params or method
	op := req.URL.Query().Get("operation")
	if op == "" {
		op = getOperation(req)
	}

	r.logger.Infow("S3 request",
		"method", req.Method,
		"bucket", bucket,
		"key", key,
		"operation", op,
	)

	switch {
	// Bucket operations
	case req.Method == http.MethodGet && bucket != "" && key == "" && op == "":
		r.ListObjectsV2(w, req, bucket)
	case req.Method == http.MethodHead && bucket != "" && key == "":
		r.HeadBucket(w, req, bucket)
	case req.Method == http.MethodPut && bucket != "" && key == "" && op == "":
		r.CreateBucket(w, req, bucket)
	case req.Method == http.MethodDelete && bucket != "" && key == "":
		r.DeleteBucket(w, req, bucket)

	// Object operations
	case req.Method == http.MethodGet && bucket != "" && key != "" && req.URL.Query().Get("attributes") != "":
		r.GetObjectAttributes(w, req, bucket, key)
	case req.Method == http.MethodGet && bucket != "" && key != "":
		r.GetObject(w, req, bucket, key)
	case req.Method == http.MethodHead && bucket != "" && key != "":
		r.HeadObject(w, req, bucket, key)
	case req.Method == http.MethodPut && bucket != "" && key != "":
		r.PutObject(w, req, bucket, key)
	case req.Method == http.MethodDelete && bucket != "" && key != "":
		r.DeleteObject(w, req, bucket, key)

	// Multipart upload
	case req.Method == http.MethodPost && bucket != "" && key != "" && op == "upload":
		r.InitiateMultipartUpload(w, req, bucket, key)
	case req.Method == http.MethodPut && bucket != "" && key != "" && req.URL.Query().Get("uploadId") != "":
		r.UploadPart(w, req, bucket, key)
	case req.Method == http.MethodPost && bucket != "" && key != "" && op == "complete":
		r.CompleteMultipartUpload(w, req, bucket, key)
	case req.Method == http.MethodDelete && bucket != "" && key != "" && req.URL.Query().Get("uploadId") != "":
		r.AbortMultipartUpload(w, req, bucket, key)
	case req.Method == http.MethodGet && bucket != "" && key != "" && op == "listparts":
		r.ListParts(w, req, bucket, key)
	case req.Method == http.MethodGet && bucket != "" && op == "listuploads":
		r.ListMultipartUploads(w, req, bucket)

	// SelectObjectContent
	case req.Method == http.MethodPost && bucket != "" && key != "" && op == "select":
		r.SelectObjectContent(w, req, bucket, key)

	// Service operations
	case req.Method == http.MethodGet && bucket == "":
		r.ListBuckets(w, req)

	// Default
	default:
		r.writeError(w, ErrNotImplemented)
	}
}

// ListBuckets lists all buckets
func (r *Router) ListBuckets(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	buckets, err := r.engine.ListBuckets(ctx)
	if err != nil {
		r.writeError(w, err)
		return
	}

	result := &s3types.ListAllMyBucketsResult{
		Owner: &s3types.Owner{
			ID:          "root",
			DisplayName: "root",
		},
		Buckets: &s3types.Buckets{},
	}

	for _, b := range buckets {
		result.Buckets.Bucket = append(result.Buckets.Bucket, s3types.Bucket{
			Name:         b.Name,
			CreationDate: time.Unix(b.CreationDate, 0).Format(time.RFC3339),
		})
	}

	r.writeXML(w, http.StatusOK, result)
}

// CreateBucket creates a new bucket
func (r *Router) CreateBucket(w http.ResponseWriter, req *http.Request, bucket string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:CreateBucket"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	if err := r.engine.CreateBucket(ctx, bucket); err != nil {
		r.writeError(w, err)
		return
	}

	s3RequestsTotal.WithLabelValues("CreateBucket", "200").Inc()
	w.WriteHeader(http.StatusOK)
}

// DeleteBucket deletes a bucket
func (r *Router) DeleteBucket(w http.ResponseWriter, req *http.Request, bucket string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:DeleteBucket"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	if err := r.engine.DeleteBucket(ctx, bucket); err != nil {
		r.writeError(w, err)
		return
	}

	s3RequestsTotal.WithLabelValues("DeleteBucket", "200").Inc()
	w.WriteHeader(http.StatusNoContent)
}

// HeadBucket checks if bucket exists
func (r *Router) HeadBucket(w http.ResponseWriter, req *http.Request, bucket string) {
	ctx := req.Context()

	buckets, err := r.engine.ListBuckets(ctx)
	if err != nil {
		r.writeError(w, err)
		return
	}

	for _, b := range buckets {
		if b.Name == bucket {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	r.writeError(w, ErrNoSuchBucket)
}

// ListObjectsV2 lists objects in a bucket
func (r *Router) ListObjectsV2(w http.ResponseWriter, req *http.Request, bucket string) {
	ctx := req.Context()

	opts := engine.ListObjectsOptions{
		Prefix:    req.URL.Query().Get("prefix"),
		Delimiter: req.URL.Query().Get("delimiter"),
		MaxKeys:   parseInt(req.URL.Query().Get("max-keys"), 1000),
		Marker:    req.URL.Query().Get("continuation-token"),
	}

	if opts.MaxKeys > 10000 {
		opts.MaxKeys = 10000
	}

	result, err := r.engine.ListObjects(ctx, bucket, opts)
	if err != nil {
		r.writeError(w, err)
		return
	}

	// Convert to S3 response
	xmlResult := &s3types.ListObjectsV2Output{
		IsTruncated: result.IsTruncated,
		Prefix:      result.Prefix,
		Delimiter:   result.Delimiter,
		MaxKeys:     strconv.Itoa(result.MaxKeys),
	}

	for _, obj := range result.Objects {
		xmlResult.Contents = append(xmlResult.Contents, s3types.Object{
			Key:          obj.Key,
			Size:         strconv.FormatInt(obj.Size, 10),
			ETag:         obj.ETag,
			LastModified: time.Unix(obj.LastModified, 0).Format(time.RFC3339),
			StorageClass: "STANDARD",
		})
	}

	if result.NextMarker != "" {
		xmlResult.NextContinuationToken = result.NextMarker
	}

	r.writeXML(w, http.StatusOK, xmlResult)
	s3RequestsTotal.WithLabelValues("ListObjectsV2", "200").Inc()
}

// PutObject stores an object
func (r *Router) PutObject(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:PutObject"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	// Get content length
	size := req.ContentLength
	if size < 0 {
		// Try to read all data
		data, err := io.ReadAll(req.Body)
		if err != nil {
			r.writeError(w, ErrInternal)
			return
		}
		req.Body = io.NopCloser(strings.NewReader(string(data)))
		size = int64(len(data))
	}

	opts := engine.PutObjectOptions{
		ContentType:     req.Header.Get("Content-Type"),
		ContentEncoding: req.Header.Get("Content-Encoding"),
		CacheControl:    req.Header.Get("Cache-Control"),
		Metadata:        parseMetadata(req),
		StorageClass:    req.Header.Get("x-amz-storage-class"),
	}

	result, err := r.engine.PutObject(ctx, bucket, key, req.Body, opts)
	if err != nil {
		r.writeError(w, err)
		return
	}

	// Set response headers
	w.Header().Set("ETag", result.ETag)
	w.Header().Set("x-amz-version-id", result.VersionID)

	s3RequestsTotal.WithLabelValues("PutObject", "200").Inc()
	w.WriteHeader(http.StatusOK)
}

// GetObject retrieves an object
func (r *Router) GetObject(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:GetObject"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	opts := engine.GetObjectOptions{
		VersionID: req.URL.Query().Get("versionId"),
		IfMatch:           req.Header.Get("If-Match"),
		IfNoneMatch:       req.Header.Get("If-None-Match"),
		IfModifiedSince:   req.Header.Get("If-Modified-Since"),
		IfUnmodifiedSince: req.Header.Get("If-Unmodified-Since"),
	}

	result, err := r.engine.GetObject(ctx, bucket, key, opts)
	if err != nil {
		r.writeError(w, err)
		return
	}
	defer result.Body.Close()

	// Set headers
	w.Header().Set("ETag", result.ETag)
	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(result.Size, 10))
	w.Header().Set("x-amz-version-id", result.VersionID)

	for k, v := range result.Metadata {
		w.Header().Set("x-amz-meta-"+k, v)
	}

	s3RequestsTotal.WithLabelValues("GetObject", "200").Inc()
	io.Copy(w, result.Body)
}

// HeadObject returns object metadata
func (r *Router) HeadObject(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:GetObject"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	info, err := r.engine.HeadObject(ctx, bucket, key)
	if err != nil {
		r.writeError(w, err)
		return
	}

	w.Header().Set("ETag", info.ETag)
	w.Header().Set("Content-Type", info.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
	w.Header().Set("x-amz-version-id", info.VersionID)

	for k, v := range info.Metadata {
		w.Header().Set("x-amz-meta-"+k, v)
	}

	s3RequestsTotal.WithLabelValues("HeadObject", "200").Inc()
	w.WriteHeader(http.StatusOK)
}

// GetObjectAttributes returns object attributes
func (r *Router) GetObjectAttributes(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:GetObjectAttributes"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	// Get version ID from query params
	versionID := req.URL.Query().Get("versionId")

	// Get object attributes
	attrs, err := r.engine.GetObjectAttributes(ctx, bucket, key, versionID)
	if err != nil {
		r.writeError(w, err)
		return
	}

	// Build response
	xmlResult := &s3types.GetObjectAttributesOutput{
		xmlns:               "http://s3.amazonaws.com/doc/2006-03-01/",
		ETag:                attrs.ETag,
		LastModified:        time.Unix(attrs.LastModified, 0).Format(time.RFC3339),
		ObjectSize:           strconv.FormatInt(attrs.Size, 10),
		StorageClass:        attrs.StorageClass,
		VersionId:           attrs.VersionID,
		ServerSideEncryption: "AES256",
	}

	// Add parts if multipart
	if len(attrs.Parts) > 0 {
		objectParts := &s3types.ObjectParts{
			TotalPartsCount: len(attrs.Parts),
			IsTruncated:     false,
		}
		for _, p := range attrs.Parts {
			objectParts.Parts = append(objectParts.Parts, s3types.Part{
				PartNumber: p.PartNumber,
				ETag:       p.ETag,
			})
		}
		xmlResult.ObjectParts = objectParts
	}

	r.writeXML(w, http.StatusOK, xmlResult)
}

// SelectObjectContent performs a select query on object data
func (r *Router) SelectObjectContent(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:GetObject"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	// Read request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.writeError(w, ErrMalformedXML)
		return
	}
	defer req.Body.Close()

	// Parse request
	var selectReq s3types.SelectObjectContentRequest
	if err := xml.Unmarshal(body, &selectReq); err != nil {
		r.writeError(w, ErrMalformedXML)
		return
	}

	// Execute select query
	result, err := r.engine.SelectObjectContent(ctx, bucket, key, selectReq.Expression)
	if err != nil {
		r.writeError(w, err)
		return
	}

	// Build response
	xmlResult := s3types.SelectObjectContentOutput{
		Payload: s3types.SelectObjectContentPayload{
			Records: &s3types.RecordsEvent{
				Body: result.Body,
			},
			Stats: &s3types.StatsEvent{
				Details: s3types.StatsDetails{
					BytesScanned:   result.BytesScanned,
					BytesProcessed: result.BytesScanned,
					BytesReturned:  result.BytesReturned,
				},
			},
		},
	}

	// Write as event stream (simplified)
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-amz-request-id", uuid.New().String())
	w.WriteHeader(http.StatusOK)

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	encoder.Encode(xmlResult)
}

// DeleteObject deletes an object
func (r *Router) DeleteObject(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:DeleteObject"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	opts := engine.DeleteObjectOptions{
		VersionID: req.URL.Query().Get("versionId"),
	}

	if err := r.engine.DeleteObject(ctx, bucket, key, opts); err != nil {
		r.writeError(w, err)
		return
	}

	s3RequestsTotal.WithLabelValues("DeleteObject", "200").Inc()
	w.WriteHeader(http.StatusNoContent)
}

// InitiateMultipartUpload initiates a multipart upload
func (r *Router) InitiateMultipartUpload(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	opts := engine.PutObjectOptions{
		ContentType: req.Header.Get("Content-Type"),
		Metadata:    parseMetadata(req),
	}

	result, err := r.engine.CreateMultipartUpload(ctx, bucket, key, opts)
	if err != nil {
		r.writeError(w, err)
		return
	}

	response := &s3types.InitiateMultipartUploadResult{
		Bucket:   result.Bucket,
		Key:      result.Key,
		UploadID: result.UploadID,
	}

	r.writeXML(w, http.StatusOK, response)
	s3RequestsTotal.WithLabelValues("InitiateMultipartUpload", "200").Inc()
}

// UploadPart uploads a part
func (r *Router) UploadPart(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	uploadID := req.URL.Query().Get("uploadId")
	partNumber := parseInt(req.URL.Query().Get("partNumber"), 0)

	result, err := r.engine.UploadPart(ctx, bucket, key, uploadID, partNumber, req.Body)
	if err != nil {
		r.writeError(w, err)
		return
	}

	w.Header().Set("ETag", result.ETag)
	s3RequestsTotal.WithLabelValues("UploadPart", "200").Inc()
	w.WriteHeader(http.StatusOK)
}

// CompleteMultipartUpload completes a multipart upload
func (r *Router) CompleteMultipartUpload(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	// Parse the request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.writeError(w, ErrMalformedXML)
		return
	}

	var complete s3types.CompleteMultipartUploadInput
	if err := xml.Unmarshal(body, &complete); err != nil {
		r.writeError(w, ErrMalformedXML)
		return
	}

	// Convert parts
	var parts []engine.PartInfo
	for _, p := range complete.Parts {
		parts = append(parts, engine.PartInfo{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		})
	}

	uploadID := req.URL.Query().Get("uploadId")
	result, err := r.engine.CompleteMultipartUpload(ctx, bucket, key, uploadID, parts)
	if err != nil {
		r.writeError(w, err)
		return
	}

	response := &s3types.CompleteMultipartUploadResult{
		Location: fmt.Sprintf("http://%s.s3.amazonaws.com/%s", bucket, key),
		Bucket:   bucket,
		Key:      key,
		ETag:     result.ETag,
	}

	r.writeXML(w, http.StatusOK, response)
	s3RequestsTotal.WithLabelValues("CompleteMultipartUpload", "200").Inc()
}

// AbortMultipartUpload aborts a multipart upload
func (r *Router) AbortMultipartUpload(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	uploadID := req.URL.Query().Get("uploadId")

	if err := r.engine.AbortMultipartUpload(ctx, bucket, key, uploadID); err != nil {
		r.writeError(w, err)
		return
	}

	s3RequestsTotal.WithLabelValues("AbortMultipartUpload", "200").Inc()
	w.WriteHeader(http.StatusNoContent)
}

// ListParts lists parts of a multipart upload
func (r *Router) ListParts(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	// Get upload ID from query parameters
	uploadID := req.URL.Query().Get("uploadId")
	if uploadID == "" {
		r.writeError(w, ErrInvalidURI)
		return
	}

	// Check bucket exists
	if _, err := r.engine.GetBucket(ctx, bucket); err != nil {
		r.writeError(w, ErrNoSuchBucket)
		return
	}

	// Get parts from engine
	parts, err := r.engine.ListParts(ctx, bucket, key, uploadID)
	if err != nil {
		r.writeError(w, err)
		return
	}

	// Build response
	xmlResult := &s3types.ListPartsOutput{
		xmlns:         "http://s3.amazonaws.com/doc/2006-03-01/",
		Bucket:        bucket,
		Key:           key,
		UploadID:      uploadID,
		StorageClass:  "STANDARD",
		IsTruncated:   false,
	}

	for _, p := range parts {
		xmlResult.Parts = append(xmlResult.Parts, s3types.Part{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		})
	}

	r.writeXML(w, http.StatusOK, xmlResult)
}

// ListMultipartUploads lists multipart uploads
func (r *Router) ListMultipartUploads(w http.ResponseWriter, req *http.Request, bucket string) {
	ctx := req.Context()

	// Check bucket exists
	if _, err := r.engine.GetBucket(ctx, bucket); err != nil {
		r.writeError(w, ErrNoSuchBucket)
		return
	}

	// Parse query parameters
	prefix := req.URL.Query().Get("prefix")
	delimiter := req.URL.Query().Get("delimiter")
	maxUploads := parseInt(req.URL.Query().Get("max-uploads"), 1000)
	keyMarker := req.URL.Query().Get("key-marker")
	uploadIDMarker := req.URL.Query().Get("upload-id-marker")

	if maxUploads > 1000 {
		maxUploads = 1000
	}

	// Get uploads from engine
	result, err := r.engine.ListMultipartUpload(ctx, bucket, prefix)
	if err != nil {
		r.writeError(w, err)
		return
	}

	// Filter by key marker if present
	var uploads []engine.MultipartUpload
	for _, u := range result.Uploads {
		if keyMarker != "" && u.Key <= keyMarker {
			continue
		}
		if uploadIDMarker != "" && u.UploadID <= uploadIDMarker && u.Key == keyMarker {
			continue
		}
		uploads = append(uploads, u)
	}

	// Apply max uploads limit
	if len(uploads) > maxUploads {
		uploads = uploads[:maxUploads]
	}

	// Build response
	xmlResult := &s3types.ListMultipartUploadsOutput{
		xmlns:      "http://s3.amazonaws.com/doc/2006-03-01/",
		Bucket:     bucket,
		MaxUploads: strconv.Itoa(maxUploads),
	}

	// Apply delimiter and build common prefixes
	if delimiter != "" {
		commonPrefixes := make(map[string]bool)
		var filteredUploads []engine.MultipartUpload

		for _, u := range uploads {
			// Find the prefix part before delimiter
			idx := strings.Index(u.Key, delimiter)
			if idx >= 0 {
				commonPrefix := u.Key[:idx+len(delimiter)]
				if !commonPrefixes[commonPrefix] {
					commonPrefixes[commonPrefix] = true
				}
			} else {
				filteredUploads = append(filteredUploads, u)
			}
		}

		for cp := range commonPrefixes {
			xmlResult.CommonPrefixes = append(xmlResult.CommonPrefixes, cp)
		}
		uploads = filteredUploads
	}

	// Convert uploads to S3 format
	for _, u := range uploads {
		xmlResult.Upload = append(xmlResult.Upload, s3types.Upload{
			Key:        u.Key,
			UploadID:  u.UploadID,
			Initiated:  time.Unix(u.Initiated, 0).Format(time.RFC3339),
			StorageClass: "STANDARD",
		})
	}

	// Set truncation
	xmlResult.IsTruncated = len(xmlResult.Upload) > maxUploads ||
		(len(xmlResult.CommonPrefixes) > 0 && len(uploads) >= maxUploads)

	if xmlResult.IsTruncated && len(xmlResult.Upload) > 0 {
		xmlResult.NextKeyMarker = xmlResult.Upload[len(xmlResult.Upload)-1].Key
		xmlResult.NextUploadIDMarker = xmlResult.Upload[len(xmlResult.Upload)-1].UploadID
	}

	// Set markers
	if keyMarker != "" {
		xmlResult.KeyMarker = keyMarker
	}
	if uploadIDMarker != "" {
		xmlResult.UploadIDMarker = uploadIDMarker
	}

	r.writeXML(w, http.StatusOK, xmlResult)
}

// writeXML writes an XML response
func (r *Router) writeXML(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	encoder.Encode(v)
}

// writeError writes an S3 error
func (r *Router) writeError(w http.ResponseWriter, err error) {
	s3Err, ok := err.(S3Error)
	if !ok {
		s3Err = ErrInternal
	}

	r.logger.Errorw("S3 error",
		"error", err,
		"code", s3Err.Code(),
	)

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(s3Err.StatusCode())

	response := s3types.Error{
		Code:      s3Err.Code(),
		Message:   s3Err.Message(),
		Resource:  "",
		RequestID: "",
	}

	encoder := xml.NewEncoder(w)
	encoder.Encode(response)
}

// parseBucketKey extracts bucket and key from the request
func parseBucketKey(req *http.Request, path string) (bucket, key string, err error) {
	// Virtual-hosted style: bucket.domain.com/key
	host := req.Host
	if strings.HasPrefix(host, "s3.") {
		host = strings.TrimPrefix(host, "s3.")
	}

	// Check for bucket in host
	if idx := strings.Index(host, "."); idx > 0 {
		bucket = host[:idx]
	} else if req.URL.Path != "" && req.URL.Path != "/" {
		// Path style: /bucket/key
		parts := strings.SplitN(strings.TrimPrefix(req.URL.Path, "/"), "/", 2)
		if len(parts) > 0 && parts[0] != "" {
			bucket = parts[0]
			if len(parts) > 1 {
				key = parts[1]
			}
		}
	} else {
		// Check query params
		bucket = req.URL.Query().Get("bucket")
	}

	return bucket, key, nil
}

// getOperation returns the S3 operation from request
func getOperation(req *http.Request) string {
	switch req.Method {
	case http.MethodGet:
		return "GetObject"
	case http.MethodPut:
		return "PutObject"
	case http.MethodDelete:
		return "DeleteObject"
	case http.MethodHead:
		return "HeadObject"
	case http.MethodPost:
		return "PostObject"
	default:
		return ""
	}
}

// parseMetadata parses x-amz-meta-* headers
func parseMetadata(req *http.Request) map[string]string {
	metadata := make(map[string]string)
	for k, v := range req.Header {
		if strings.HasPrefix(k, "x-amz-meta-") {
			key := strings.TrimPrefix(k, "x-amz-meta-")
			if len(v) > 0 {
				metadata[key] = v[0]
			}
		}
	}
	return metadata
}

// parseInt parses an integer with default
func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return i
}
