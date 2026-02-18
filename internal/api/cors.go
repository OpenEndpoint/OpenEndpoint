package api

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/openendpoint/openendpoint/internal/auth"
	"github.com/openendpoint/openendpoint/internal/config"
	"github.com/openendpoint/openendpoint/internal/engine"
	"github.com/openendpoint/openendpoint/internal/metadata"
	s3types "github.com/openendpoint/openendpoint/pkg/s3types"
	"go.uber.org/zap"
)

// CORSConfig holds CORS configuration for a bucket
type CORSConfig struct {
	Rules []CORSRule
}

// CORSRule represents a CORS rule
type CORSRule struct {
	ID             string
	AllowedMethods []string
	AllowedOrigins []string
	AllowedHeaders []string
	ExposeHeaders  []string
	MaxAgeSeconds int
}

// Router with CORS support
type Router struct {
	engine      *engine.ObjectService
	auth        *auth.Auth
	logger      *zap.SugaredLogger
	config      *config.Config
	corsRules   map[string]*CORSConfig
}

// NewRouter creates a new S3 API router with CORS
func NewRouter(engine *engine.ObjectService, auth *auth.Auth, logger *zap.SugaredLogger, cfg *config.Config) *Router {
	return &Router{
		engine:    engine,
		auth:      auth,
		logger:    logger,
		config:    cfg,
		corsRules: make(map[string]*CORSConfig),
	}
}

// ServeHTTP handles S3 API requests with CORS
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Parse bucket and key
	bucket, key, err := parseBucketKey(req, req.URL.Path)
	if err != nil {
		r.writeError(w, ErrInvalidURI)
		return
	}

	// Apply CORS headers if this is a bucket request
	if bucket != "" {
		r.applyCORS(w, req, bucket)
	}

	// Handle OPTIONS for CORS preflight
	if req.Method == http.MethodOptions {
		r.handleCORSpreflight(w, req, bucket)
		return
	}

	// Continue with normal routing
	r.route(w, req, bucket, key)
}

// applyCORS applies CORS headers to the response
func (r *Router) applyCORS(w http.ResponseWriter, req *http.Request, bucket string) {
	cors, ok := r.corsRules[bucket]
	if !ok {
		// Use default CORS
		cors = &CORSConfig{
			Rules: []CORSRule{
				{
					AllowedOrigins: []string{"*"},
					AllowedMethods: []string{"GET", "PUT", "DELETE", "HEAD", "POST"},
					AllowedHeaders: []string{"*"},
				},
			},
		}
	}

	origin := req.Header.Get("Origin")
	if origin == "" {
		return
	}

	// Check if origin is allowed
	for _, rule := range cors.Rules {
		if matchesOrigin(origin, rule.AllowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(rule.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(rule.AllowedHeaders, ", "))
			if len(rule.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(rule.ExposeHeaders, ", "))
			}
			if rule.MaxAgeSeconds > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(rule.MaxAgeSeconds))
			}
			break
		}
	}
}

// handleCORSpreflight handles CORS preflight requests
func (r *Router) handleCORSpreflight(w http.ResponseWriter, req *http.Request, bucket string) {
	origin := req.Header.Get("Origin")
	accessControlRequestMethod := req.Header.Get("Access-Control-Request-Method")

	if origin == "" || accessControlRequestMethod == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Apply CORS headers
	r.applyCORS(w, req, bucket)

	w.WriteHeader(http.StatusOK)
}

// matchesOrigin checks if an origin matches allowed origins
func matchesOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
		if strings.HasPrefix(allowed, "*.") {
			suffix := allowed[2:]
			if strings.HasSuffix(origin, suffix) {
				return true
			}
		}
	}
	return false
}

// CopyObject copies an object
func (r *Router) CopyObject(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:PutObject"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	// Get copy source from header
	copySource := req.Header.Get("x-amz-copy-source")
	if copySource == "" {
		r.writeError(w, ErrInvalidArgument)
		return
	}

	// Parse source (format: /bucket/key or bucket/key)
	copySource = strings.TrimPrefix(copySource, "/")
	srcParts := strings.SplitN(copySource, "/", 2)
	if len(srcParts) != 2 {
		r.writeError(w, ErrInvalidURI)
		return
	}

	srcBucket := srcParts[0]
	srcKey := srcParts[1]

	// Get source object
	result, err := r.engine.GetObject(ctx, srcBucket, srcKey, engine.GetObjectOptions{})
	if err != nil {
		r.writeError(w, err)
		return
	}
	defer result.Body.Close()

	// Copy to destination
	dstOpts := engine.PutObjectOptions{
		ContentType: result.ContentType,
		Metadata:    result.Metadata,
	}

	copyResult, err := r.engine.PutObject(ctx, bucket, key, result.Body, dstOpts)
	if err != nil {
		r.writeError(w, err)
		return
	}

	// Build response
	response := &s3types.CopyObjectResult{
		LastModified: time.Now().Format(time.RFC3339),
		ETag:        copyResult.ETag,
		RequestID:   "",
	}

	r.writeXML(w, http.StatusOK, response)
	s3RequestsTotal.WithLabelValues("CopyObject", "200").Inc()
}

// GetObjectAttributes returns object attributes
func (r *Router) GetObjectAttributes(w http.ResponseWriter, req *http.Request, bucket, key string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:GetObject"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	// Get object info
	info, err := r.engine.HeadObject(ctx, bucket, key)
	if err != nil {
		r.writeError(w, err)
		return
	}

	// Build response
	// Note: This is a simplified version
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("ETag", info.ETag)
	w.Header().Set("x-amz-version-id", info.VersionID)

	// Write object attributes
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<GetObjectAttributesResponse xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <ETag>%s</ETag>
  <ObjectSize>%d</ObjectSize>
  <LastModified>%s</LastModified>
  <VersionId>%s</VersionId>
</GetObjectAttributesResponse>`, info.ETag, info.Size, time.Unix(info.LastModified, 0).Format(time.RFC3339), info.VersionID)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xml))
}

// PutBucketCors creates CORS configuration
func (r *Router) PutBucketCors(w http.ResponseWriter, req *http.Request, bucket string) {
	ctx := req.Context()

	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:PutBucketCORS"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	// Parse CORS configuration
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.writeError(w, ErrMalformedXML)
		return
	}

	var corsReq struct {
		XMLName xml.Name `xml:"CORSConfiguration"`
		Rules   []struct {
			XMLName             xml.Name `xml:"CORSRule"`
			ID                  string   `xml:"ID"`
			AllowedMethods      []string `xml:"AllowedMethod"`
			AllowedOrigins      []string `xml:"AllowedOrigin"`
			AllowedHeaders      []string `xml:"AllowedHeader"`
			ExposeHeaders       []string `xml:"ExposeHeader"`
			MaxAgeSeconds       int      `xml:"MaxAgeSeconds"`
		} `xml:"CORSRule"`
	}

	if err := xml.Unmarshal(body, &corsReq); err != nil {
		r.writeError(w, ErrMalformedXML)
		return
	}

	// Store CORS rules
	var rules []CORSRule
	for _, r := range corsReq.Rules {
		rules = append(rules, CORSRule{
			ID:              r.ID,
			AllowedMethods:  r.AllowedMethods,
			AllowedOrigins:  r.AllowedOrigins,
			AllowedHeaders:  r.AllowedHeaders,
			ExposeHeaders:   r.ExposeHeaders,
			MaxAgeSeconds:   r.MaxAgeSeconds,
		})
	}

	r.corsRules[bucket] = &CORSConfig{Rules: rules}

	_ = ctx // Suppress unused warning

	s3RequestsTotal.WithLabelValues("PutBucketCors", "200").Inc()
	w.WriteHeader(http.StatusOK)
}

// GetBucketCors returns CORS configuration
func (r *Router) GetBucketCors(w http.ResponseWriter, req *http.Request, bucket string) {
	cors, ok := r.corsRules[bucket]

	w.Header().Set("Content-Type", "application/xml")

	if !ok || len(cors.Rules) == 0 {
		// Return default or empty
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<CORSConfiguration/>`))
		return
	}

	// Build XML response
	var xml strings.Builder
	xml.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<CORSConfiguration>`)

	for _, rule := range cors.Rules {
		xml.WriteString(`<CORSRule>`)
		if rule.ID != "" {
			xml.WriteString(`<ID>` + rule.ID + `</ID>`)
		}
		for _, origin := range rule.AllowedOrigins {
			xml.WriteString(`<AllowedOrigin>` + origin + `</AllowedOrigin>`)
		}
		for _, method := range rule.AllowedMethods {
			xml.WriteString(`<AllowedMethod>` + method + `</AllowedMethod>`)
		}
		for _, header := range rule.AllowedHeaders {
			xml.WriteString(`<AllowedHeader>` + header + `</AllowedHeader>`)
		}
		xml.WriteString(`</CORSRule>`)
	}

	xml.WriteString(`</CORSConfiguration>`)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xml.String()))
}

// DeleteBucketCors deletes CORS configuration
func (r *Router) DeleteBucketCors(w http.ResponseWriter, req *http.Request, bucket string) {
	delete(r.corsRules, bucket)
	s3RequestsTotal.WithLabelValues("DeleteBucketCors", "200").Inc()
	w.WriteHeader(http.StatusNoContent)
}

// PutBucketPolicy creates bucket policy
func (r *Router) PutBucketPolicy(w http.ResponseWriter, req *http.Request, bucket string) {
	// Check authorization
	if err := r.auth.Authorize(req, bucket, "s3:PutBucketPolicy"); err != nil {
		r.writeError(w, ErrAccessDenied)
		return
	}

	// Read policy
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.writeError(w, ErrMalformedXML)
		return
	}

	// Validate JSON
	if !isValidJSON(body) {
		r.writeError(w, ErrMalformedJSON)
		return
	}

	// Store policy (simplified - in production, use proper policy storage)
	_ = body // Store policy

	s3RequestsTotal.WithLabelValues("PutBucketPolicy", "200").Inc()
	w.WriteHeader(http.StatusNoContent)
}

// GetBucketPolicy returns bucket policy
func (r *Router) GetBucketPolicy(w http.ResponseWriter, req *http.Request, bucket string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}")) // Return empty policy for now
}

// DeleteBucketPolicy deletes bucket policy
func (r *Router) DeleteBucketPolicy(w http.ResponseWriter, req *http.Request, bucket string) {
	s3RequestsTotal.WithLabelValues("DeleteBucketPolicy", "200").Inc()
	w.WriteHeader(http.StatusNoContent)
}

// isValidJSON checks if a string is valid JSON
func isValidJSON(data []byte) bool {
	var js interface{}
	return json.Unmarshal(data, &js) == nil
}

// GetBucketLocation returns bucket location
func (r *Router) GetBucketLocation(w http.ResponseWriter, req *http.Request, bucket string) {
	response := &s3types.GetBucketLocationOutput{
		Region: "us-east-1",
	}

	r.writeXML(w, http.StatusOK, response)
}

// GetBucketVersioning returns bucket versioning status
func (r *Router) GetBucketVersioning(w http.ResponseWriter, req *http.Request, bucket string) {
	ctx := req.Context()

	versioning, err := r.engine.GetBucketVersioning(ctx, bucket)
	if err != nil || versioning == nil {
		versioning = &metadata.BucketVersioning{
			Status: "",
		}
	}

	response := &s3types.GetBucketVersioningOutput{
		Status:    versioning.Status,
		MFADelete: versioning.MFADelete,
	}

	r.writeXML(w, http.StatusOK, response)
}

// PutBucketVersioning sets bucket versioning
func (r *Router) PutBucketVersioning(w http.ResponseWriter, req *http.Request, bucket string) {
	ctx := req.Context()

	// Parse request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.writeError(w, ErrMalformedXML)
		return
	}

	var input s3types.PutBucketVersioningInput
	if err := xml.Unmarshal(body, &input); err != nil {
		r.writeError(w, ErrMalformedXML)
		return
	}

	versioning := &metadata.BucketVersioning{
		Status:    input.Status,
		MFADelete: input.MFADelete,
	}

	err = r.engine.PutBucketVersioning(ctx, bucket, versioning)
	if err != nil {
		r.writeError(w, err)
		return
	}

	s3RequestsTotal.WithLabelValues("PutBucketVersioning", "200").Inc()
	w.WriteHeader(http.StatusOK)
}

// Import metadata package
func init() {
	_ = storage.Range{} // Suppress unused import
}
