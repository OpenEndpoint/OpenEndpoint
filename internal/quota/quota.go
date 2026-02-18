package quota

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// QuotaType represents the type of quota
type QuotaType string

const (
	QuotaTypeStorage QuotaType = "storage" // Storage in bytes
	QuotaTypeObjects QuotaType = "objects" // Number of objects
	QuotaTypeBandwidth QuotaType = "bandwidth" // Bandwidth in bytes per second
)

// Quota represents a bucket quota
type Quota struct {
	Bucket         string
	Type           QuotaType
	Limit          int64
	Used           int64
	WarningThreshold float64 // 0.0 to 1.0
	Enforce        bool
	LastUpdated    time.Time
}

// QuotaManager manages bucket quotas
type QuotaManager struct {
	mu      sync.RWMutex
	quotas  map[string]*Quota
	bandwidth map[string]*BandwidthTracker
}

// BandwidthTracker tracks bandwidth usage
type BandwidthTracker struct {
	mu           sync.RWMutex
	readBytes    int64
	writeBytes   int64
	readOps     int64
	writeOps    int64
	lastReset    time.Time
	limit        int64 // bytes per second
}

// NewQuotaManager creates a new quota manager
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotas:    make(map[string]*Quota),
		bandwidth: make(map[string]*BandwidthTracker),
	}
}

// SetQuota sets quota for a bucket
func (qm *QuotaManager) SetQuota(bucket string, quotaType QuotaType, limit int64, warningThreshold float64, enforce bool) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quota := &Quota{
		Bucket:          bucket,
		Type:            quotaType,
		Limit:           limit,
		WarningThreshold: warningThreshold,
		Enforce:         enforce,
		LastUpdated:     time.Now(),
	}

	// Get current usage
	quota.Used = qm.getCurrentUsage(bucket, quotaType)

	qm.quotas[bucket] = quota
	return nil
}

// GetQuota gets quota for a bucket
func (qm *QuotaManager) GetQuota(bucket string) (*Quota, bool) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quota, ok := qm.quotas[bucket]
	return quota, ok
}

// DeleteQuota deletes quota for a bucket
func (qm *QuotaManager) DeleteQuota(bucket string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	delete(qm.quotas, bucket)
}

// CheckQuota checks if operation is within quota
func (qm *QuotaManager) CheckQuota(ctx context.Context, bucket string, quotaType QuotaType, size int64) (bool, string, error) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quota, ok := qm.quotas[bucket]
	if !ok {
		return true, "", nil // No quota set
	}

	if quota.Type != quotaType {
		return true, "", nil // Different quota type
	}

	newUsage := quota.Used + size

	if quota.Enforce && newUsage > quota.Limit {
		return false, "QuotaExceeded", fmt.Errorf("quota exceeded: %d/%d", newUsage, quota.Limit)
	}

	// Check warning threshold
	if quota.WarningThreshold > 0 {
		usagePercent := float64(newUsage) / float64(quota.Limit)
		if usagePercent >= quota.WarningThreshold {
			return true, "QuotaWarning", nil
		}
	}

	return true, "", nil
}

// UpdateUsage updates quota usage
func (qm *QuotaManager) UpdateUsage(bucket string, quotaType QuotaType, size int64) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quota, ok := qm.quotas[bucket]
	if !ok {
		return nil // No quota set
	}

	if quota.Type == quotaType {
		quota.Used += size
		quota.LastUpdated = time.Now()
	}

	return nil
}

// ResetUsage resets quota usage
func (qm *QuotaManager) ResetUsage(bucket string) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quota, ok := qm.quotas[bucket]
	if !ok {
		return nil
	}

	quota.Used = 0
	quota.LastUpdated = time.Now()

	return nil
}

// SetBandwidthLimit sets bandwidth limit for a bucket
func (qm *QuotaManager) SetBandwidthLimit(bucket string, limit int64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.bandwidth[bucket] = &BandwidthTracker{
		limit:     limit,
		lastReset: time.Now(),
	}
}

// CheckBandwidth checks if bandwidth is within limit
func (qm *QuotaManager) CheckBandwidth(bucket string, read, write int64) (bool, error) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	tracker, ok := qm.bandwidth[bucket]
	if !ok || tracker.limit == 0 {
		return true, nil // No limit set
	}

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Reset if a second has passed
	now := time.Now()
	if now.Sub(tracker.lastReset) >= time.Second {
		tracker.readBytes = 0
		tracker.writeBytes = 0
		tracker.lastReset = now
	}

	newRead := tracker.readBytes + read
	newWrite := tracker.writeBytes + write

	if newRead > tracker.limit || newWrite > tracker.limit {
		return false, fmt.Errorf("bandwidth limit exceeded: read=%d/%d, write=%d/%d",
			newRead, tracker.limit, newWrite, tracker.limit)
	}

	tracker.readBytes = newRead
	tracker.writeBytes = newWrite

	return true, nil
}

// GetUsage gets current quota usage
func (qm *QuotaManager) GetUsage(bucket string) (map[string]interface{}, error) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quota, ok := qm.quotas[bucket]
	if !ok {
		return nil, fmt.Errorf("quota not set for bucket: %s", bucket)
	}

	usagePercent := 0.0
	if quota.Limit > 0 {
		usagePercent = float64(quota.Used) / float64(quota.Limit) * 100
	}

	return map[string]interface{}{
		"bucket":        bucket,
		"type":         quota.Type,
		"limit":        quota.Limit,
		"used":         quota.Used,
		"available":    quota.Limit - quota.Used,
		"usage_percent": usagePercent,
		"warning_threshold": quota.WarningThreshold * 100,
		"enforce":      quota.Enforce,
		"last_updated": quota.LastUpdated,
	}, nil
}

// ListQuotas lists all quotas
func (qm *QuotaManager) ListQuotas() []*Quota {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quotas := make([]*Quota, 0, len(qm.quotas))
	for _, quota := range qm.quotas {
		quotas = append(quotas, quota)
	}

	return quotas
}

// getCurrentUsage gets current usage for a bucket (placeholder - would query storage)
func (qm *QuotaManager) getCurrentUsage(bucket string, quotaType QuotaType) int64 {
	// This would query the actual storage to get current usage
	// For now, return 0
	return 0
}

// ComplianceChecker checks compliance requirements
type ComplianceChecker struct {
	quotaManager *QuotaManager
}

// NewComplianceChecker creates a new compliance checker
func NewComplianceChecker(qm *QuotaManager) *ComplianceChecker {
	return &ComplianceChecker{
		quotaManager: qm,
	}
}

// CheckCompliance checks if bucket meets compliance requirements
func (cc *ComplianceChecker) CheckCompliance(bucket string) (map[string]bool, error) {
	result := make(map[string]bool)

	// Check if quota is set
	_, hasQuota := cc.quotaManager.quotas[bucket]
	result["has_quota"] = hasQuota

	// Check usage
	if quota, ok := cc.quotaManager.quotas[bucket]; ok {
		result["within_quota"] = quota.Used <= quota.Limit
		result["has_warning_threshold"] = quota.WarningThreshold > 0
		result["enforced"] = quota.Enforce
	}

	// Check bandwidth
	_, hasBandwidth := cc.quotaManager.bandwidth[bucket]
	result["has_bandwidth_limit"] = hasBandwidth

	return result, nil
}
