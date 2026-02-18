package analytics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// StorageMetrics contains storage metrics
type StorageMetrics struct {
	TotalBytes    int64            `json:"total_bytes"`
	TotalObjects  int64            `json:"total_objects"`
	TotalBuckets int              `json:"total_buckets"`
	ByBucket     map[string]BucketMetrics `json:"by_bucket"`
	ByStorageClass map[string]int64 `json:"by_storage_class"`
}

// BucketMetrics contains per-bucket metrics
type BucketMetrics struct {
	Name          string  `json:"name"`
	Bytes         int64   `json:"bytes"`
	Objects       int64   `json:"objects"`
	AvgObjectSize float64 `json:"avg_object_size"`
}

// RequestMetrics contains request metrics
type RequestMetrics struct {
	TotalRequests   int64   `json:"total_requests"`
	TotalErrors     int64   `json:"total_errors"`
	BytesUploaded   int64   `json:"bytes_uploaded"`
	BytesDownloaded int64   `json:"bytes_downloaded"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
}

// AccessPattern contains access pattern data
type AccessPattern struct {
	Bucket       string    `json:"bucket"`
	Key          string    `json:"key"`
	AccessCount  int64     `json:"access_count"`
	LastAccess   time.Time `json:"last_access"`
	AccessFrequency float64 `json:"access_frequency"` // accesses per hour
}

// Report contains analytics report data
type Report struct {
	Period        ReportPeriod `json:"period"`
	GeneratedAt   time.Time    `json:"generated_at"`
	Storage       StorageMetrics `json:"storage"`
	Requests      RequestMetrics `json:"requests"`
	TopBuckets    []BucketMetrics `json:"top_buckets"`
	TopObjects    []ObjectAccess `json:"top_objects"`
	CostEstimate  CostEstimate `json:"cost_estimate"`
}

// ReportPeriod represents a report time period
type ReportPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ObjectAccess contains object access info
type ObjectAccess struct {
	Bucket     string `json:"bucket"`
	Key        string `json:"key"`
	Size       int64  `json:"size"`
	Accesses   int64  `json:"accesses"`
	TotalBytes int64  `json:"total_bytes"`
}

// CostEstimate contains cost estimation
type CostEstimate struct {
	StorageCost    float64            `json:"storage_cost"`
	RequestsCost   float64            `json:"requests_cost"`
	BandwidthCost float64            `json:"bandwidth_cost"`
	TotalCost     float64            `json:"total_cost"`
	ByTier        map[string]float64 `json:"by_tier"`
}

// MetricsCollector collects storage and request metrics
type MetricsCollector struct {
	logger *zap.Logger
	mu     sync.RWMutex

	// Storage metrics
	totalBytes    int64
	totalObjects  int64
	totalBuckets  int
	bucketMetrics map[string]*BucketMetrics
	storageClass  map[string]int64

	// Request metrics
	totalRequests   int64
	totalErrors     int64
	bytesUploaded   int64
	bytesDownloaded int64
	latencySum     float64
	latencyCount   int64

	// Access tracking
	accessPatterns map[string]*AccessPattern
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:        logger,
		bucketMetrics: make(map[string]*BucketMetrics),
		storageClass:  make(map[string]int64),
		accessPatterns: make(map[string]*AccessPattern),
	}
}

// RecordObject stores object storage metrics
func (m *MetricsCollector) RecordObject(bucket, key string, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update bucket metrics
	bm, ok := m.bucketMetrics[bucket]
	if !ok {
		bm = &BucketMetrics{Name: bucket}
		m.bucketMetrics[bucket] = bm
		m.totalBuckets++
	}

	bm.Bytes += size
	bm.Objects++
	bm.AvgObjectSize = float64(bm.Bytes) / float64(bm.Objects)

	// Update totals
	m.totalBytes += size
	m.totalObjects++
}

// DeleteObject removes object metrics
func (m *MetricsCollector) DeleteObject(bucket, key string, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update bucket metrics
	if bm, ok := m.bucketMetrics[bucket]; ok {
		bm.Bytes -= size
		bm.Objects--
		if bm.Objects > 0 {
			bm.AvgObjectSize = float64(bm.Bytes) / float64(bm.Objects)
		}
	}

	// Update totals
	m.totalBytes -= size
	m.totalObjects--
}

// RecordRequest records request metrics
func (m *MetricsCollector) RecordRequest(op string, success bool, bytes int64, latencyMs float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests++
	m.latencySum += latencyMs
	m.latencyCount++

	if !success {
		m.totalErrors++
	}

	switch op {
	case "PutObject", "CopyObject", "UploadPart":
		m.bytesUploaded += bytes
	case "GetObject":
		m.bytesDownloaded += bytes
	}
}

// RecordAccess records object access
func (m *MetricsCollector) RecordAccess(bucket, key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	objKey := bucket + "/" + key
	pattern, ok := m.accessPatterns[objKey]
	if !ok {
		pattern = &AccessPattern{
			Bucket:     bucket,
			Key:        key,
			AccessCount: 0,
			LastAccess:  time.Now(),
		}
		m.accessPatterns[objKey] = pattern
	}

	pattern.AccessCount++
	pattern.LastAccess = time.Now()

	// Calculate frequency (accesses per hour since creation)
	age := time.Since(time.Unix(0, 0)) // Would use created time
	if age.Hours() > 0 {
		pattern.AccessFrequency = float64(pattern.AccessCount) / age.Hours()
	}
}

// GetStorageMetrics returns current storage metrics
func (m *MetricsCollector) GetStorageMetrics() StorageMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	byBucket := make(map[string]BucketMetrics)
	for k, v := range m.bucketMetrics {
		byBucket[k] = *v
	}

	return StorageMetrics{
		TotalBytes:     m.totalBytes,
		TotalObjects:   m.totalObjects,
		TotalBuckets:   m.totalBuckets,
		ByBucket:       byBucket,
		ByStorageClass: m.storageClass,
	}
}

// GetRequestMetrics returns current request metrics
func (m *MetricsCollector) GetRequestMetrics() RequestMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgLatency := 0.0
	if m.latencyCount > 0 {
		avgLatency = m.latencySum / float64(m.latencyCount)
	}

	return RequestMetrics{
		TotalRequests:    m.totalRequests,
		TotalErrors:     m.totalErrors,
		BytesUploaded:   m.bytesUploaded,
		BytesDownloaded: m.bytesDownloaded,
		AvgLatencyMs:   avgLatency,
	}
}

// GenerateReport generates an analytics report
func (m *MetricsCollector) GenerateReport(ctx context.Context, start, end time.Time) *Report {
	storage := m.GetStorageMetrics()
	requests := m.GetRequestMetrics()

	// Get top buckets by size
	topBuckets := m.getTopBuckets(10)

	// Get top objects by access
	topObjects := m.getTopObjects(10)

	// Calculate costs
	cost := m.calculateCosts()

	return &Report{
		Period: ReportPeriod{
			Start: start,
			End:   end,
		},
		GeneratedAt:  time.Now(),
		Storage:      storage,
		Requests:     requests,
		TopBuckets:   topBuckets,
		TopObjects:   topObjects,
		CostEstimate: cost,
	}
}

// getTopBuckets returns top buckets by size
func (m *MetricsCollector) getTopBuckets(limit int) []BucketMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	buckets := make([]BucketMetrics, 0, len(m.bucketMetrics))
	for _, bm := range m.bucketMetrics {
		buckets = append(buckets, *bm)
	}

	// Sort by bytes descending
	for i := 0; i < len(buckets); i++ {
		for j := i + 1; j < len(buckets); j++ {
			if buckets[j].Bytes > buckets[i].Bytes {
				buckets[i], buckets[j] = buckets[j], buckets[i]
			}
		}
	}

	if len(buckets) > limit {
		return buckets[:limit]
	}
	return buckets
}

// getTopObjects returns top objects by access count
func (m *MetricsCollector) getTopObjects(limit int) []ObjectAccess {
	m.mu.RLock()
	defer m.mu.RUnlock()

	objects := make([]ObjectAccess, 0, len(m.accessPatterns))
	for _, pattern := range m.accessPatterns {
		objects = append(objects, ObjectAccess{
			Bucket:     pattern.Bucket,
			Key:        pattern.Key,
			Accesses:   pattern.AccessCount,
			TotalBytes: 0, // Would need to look up
		})
	}

	// Sort by accesses descending
	for i := 0; i < len(objects); i++ {
		for j := i + 1; j < len(objects); j++ {
			if objects[j].Accesses > objects[i].Accesses {
				objects[i], objects[j] = objects[j], objects[i]
			}
		}
	}

	if len(objects) > limit {
		return objects[:limit]
	}
	return objects
}

// calculateCosts estimates costs
func (m *MetricsCollector) calculateCosts() CostEstimate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Simplified pricing (would be configurable)
	costPerGBMonth := 0.023
	requestCostPer10k := 0.0004
	bandwidthPerGB := 0.09

	storageCost := float64(m.totalBytes) / (1024 * 1024 * 1024) * costPerGBMonth
	requestsCost := float64(m.totalRequests) / 10000 * requestCostPer10k
	bandwidthCost := float64(m.bytesUploaded+m.bytesDownloaded) / (1024 * 1024 * 1024) * bandwidthPerGB

	return CostEstimate{
		StorageCost:    storageCost,
		RequestsCost:   requestsCost,
		BandwidthCost:  bandwidthCost,
		TotalCost:      storageCost + requestsCost + bandwidthCost,
		ByTier: map[string]float64{
			"hot": storageCost * 0.8,
			"cold": storageCost * 0.2,
		},
	}
}

// Reporter provides analytics reporting
type Reporter struct {
	collector *MetricsCollector
	logger    *zap.Logger
}

// NewReporter creates a new reporter
func NewReporter(collector *MetricsCollector, logger *zap.Logger) *Reporter {
	return &Reporter{
		collector: collector,
		logger:    logger,
	}
}

// GenerateHourlyReport generates hourly report
func (r *Reporter) GenerateHourlyReport(ctx context.Context) (*Report, error) {
	end := time.Now()
	start := end.Add(-1 * time.Hour)

	report := r.collector.GenerateReport(ctx, start, end)

	r.logger.Info("Hourly report generated",
		zap.Int64("total_bytes", report.Storage.TotalBytes),
		zap.Int64("total_requests", report.Requests.TotalRequests),
		zap.Float64("total_cost", report.CostEstimate.TotalCost))

	return report, nil
}

// GenerateDailyReport generates daily report
func (r *Reporter) GenerateDailyReport(ctx context.Context) (*Report, error) {
	end := time.Now()
	start := end.Add(-24 * time.Hour)

	return r.collector.GenerateReport(ctx, start, end), nil
}

// GenerateMonthlyReport generates monthly report
func (r *Reporter) GenerateMonthlyReport(ctx context.Context) (*Report, error) {
	end := time.Now()
	start := end.Add(-30 * 24 * time.Hour)

	return r.collector.GenerateReport(ctx, start, end), nil
}

// PredictGrowth predicts storage growth
func (r *Reporter) PredictGrowth(ctx context.Context, days int) (int64, error) {
	// Get historical data
	storage := r.collector.GetStorageMetrics()
	requests := r.collector.GetRequestMetrics()

	// Simple linear projection based on recent growth
	// In production, use more sophisticated forecasting
	avgObjectSize := float64(0)
	if storage.TotalObjects > 0 {
		avgObjectSize = float64(storage.TotalBytes) / float64(storage.TotalObjects)
	}

	// Assume uploads continue at current rate
	objectsPerHour := float64(requests.TotalRequests) / 24 // rough estimate
	objectsPerDay := objectsPerHour * 24
	predictedGrowth := int64(objectsPerDay * avgObjectSize * float64(days))

	r.logger.Info("Growth prediction",
		zap.Int("days", days),
		zap.Int64("predicted_bytes", predictedGrowth))

	return predictedGrowth, nil
}

// GetInsights returns actionable insights
func (r *Reporter) GetInsights(ctx context.Context) []Insight {
	var insights []Insight

	storage := r.collector.GetStorageMetrics()
	requests := r.collector.GetRequestMetrics()

	// Check for high error rate
	if requests.TotalRequests > 0 {
		errorRate := float64(requests.TotalErrors) / float64(requests.TotalRequests) * 100
		if errorRate > 5 {
			insights = append(insights, Insight{
				Type:    "error_rate",
				Level:   "warning",
				Message: fmt.Sprintf("Error rate is %.2f%% - consider investigating", errorRate),
			})
		}
	}

	// Check for large objects
	if storage.TotalObjects > 0 {
		avgSize := float64(storage.TotalBytes) / float64(storage.TotalObjects)
		if avgSize > 100*1024*1024 { // > 100MB
			insights = append(insights, Insight{
				Type:    "large_objects",
				Level:   "info",
				Message: fmt.Sprintf("Average object size is %.2f MB - consider using multipart upload", avgSize/1024/1024),
			})
		}
	}

	// Check for bucket concentration
	if len(storage.ByBucket) > 0 {
		var largestBucket string
		var largestSize int64
		for name, metrics := range storage.ByBucket {
			if metrics.Bytes > largestSize {
				largestSize = metrics.Bytes
				largestBucket = name
			}
		}

		if largestSize > storage.TotalBytes*80/100 {
			insights = append(insights, Insight{
				Type:    "bucket_concentration",
				Level:   "info",
				Message: fmt.Sprintf("Bucket '%s' contains 80%% of data - consider better distribution", largestBucket),
			})
		}
	}

	return insights
}

// Insight represents an actionable insight
type Insight struct {
	Type    string `json:"type"`
	Level   string `json:"level"` // info, warning, error
	Message string `json:"message"`
}
