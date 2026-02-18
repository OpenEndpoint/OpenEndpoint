package locking

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LockType represents the type of lock
type LockType string

const (
	LockTypeExclusive LockType = "exclusive"
	LockTypeShared    LockType = "shared"
)

// ObjectLock handles object locking for compliance/retention
type ObjectLock struct {
	mu          sync.RWMutex
	locks       map[string]*ObjectLockConfig
	retention   map[string]*RetentionPolicy
	legalHolds  map[string]*LegalHold
}

// ObjectLockConfig holds lock configuration
type ObjectLockConfig struct {
	Enabled              bool   `json:"enabled"`
	RetentionMode        string `json:"retention_mode"` // GOVERNANCE or COMPLIANCE
	RetentionDays        int    `json:"retention_days"`
	RetentionYears       int    `json:"retention_years"`
}

// RetentionPolicy represents object retention policy
type RetentionPolicy struct {
	Bucket       string
	Key          string
	Mode         string // GOVERNANCE or COMPLIANCE
	RetainUntil time.Time
	CreatedAt    time.Time
	CreatedBy    string
}

// LegalHold represents a legal hold
type LegalHold struct {
	Bucket     string
	Key        string
	Status     string // ON or OFF
	CreatedAt  time.Time
	CreatedBy  string
}

// NewObjectLock creates a new object lock manager
func NewObjectLock() *ObjectLock {
	return &ObjectLock{
		locks:      make(map[string]*ObjectLockConfig),
		retention:  make(map[string]*RetentionPolicy),
		legalHolds: make(map[string]*LegalHold),
	}
}

// EnableObjectLock enables object locking for a bucket
func (ol *ObjectLock) EnableObjectLock(bucket string, mode string, days, years int) error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	ol.locks[bucket] = &ObjectLockConfig{
		Enabled:        true,
		RetentionMode:  mode,
		RetentionDays:  days,
		RetentionYears: years,
	}

	return nil
}

// DisableObjectLock disables object locking for a bucket
func (ol *ObjectLock) DisableObjectLock(bucket string) error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	delete(ol.locks, bucket)
	return nil
}

// IsObjectLockEnabled checks if object lock is enabled for a bucket
func (ol *ObjectLock) IsObjectLockEnabled(bucket string) bool {
	ol.mu.RLock()
	defer ol.mu.RUnlock()

	lock, ok := ol.locks[bucket]
	return ok && lock.Enabled
}

// GetLockConfig gets object lock configuration for a bucket
func (ol *ObjectLock) GetLockConfig(bucket string) (*ObjectLockConfig, bool) {
	ol.mu.RLock()
	defer ol.mu.RUnlock()

	lock, ok := ol.locks[bucket]
	return lock, ok
}

// SetRetention sets retention policy for an object
func (ol *ObjectLock) SetRetention(ctx context.Context, bucket, key, mode string, retainUntil time.Time, createdBy string) error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	// Check if object lock is enabled
	lock, ok := ol.locks[bucket]
	if !ok || !lock.Enabled {
		return fmt.Errorf("object lock not enabled for bucket: %s", bucket)
	}

	// Check retention mode
	if mode != "GOVERNANCE" && mode != "COMPLIANCE" {
		return fmt.Errorf("invalid retention mode: %s", mode)
	}

	// For COMPLIANCE mode, retention cannot be reduced or removed
	if lock.RetentionMode == "COMPLIANCE" {
		existing, exists := ol.retention[fmt.Sprintf("%s/%s", bucket, key)]
		if exists && existing.Mode == "COMPLIANCE" {
			if retainUntil.Before(existing.RetainUntil) {
				return fmt.Errorf("cannot reduce retention period for COMPLIANCE locked object")
			}
		}
	}

	keyStr := fmt.Sprintf("%s/%s", bucket, key)
	ol.retention[keyStr] = &RetentionPolicy{
		Bucket:       bucket,
		Key:          key,
		Mode:         mode,
		RetainUntil:  retainUntil,
		CreatedAt:    time.Now(),
		CreatedBy:    createdBy,
	}

	return nil
}

// GetRetention gets retention policy for an object
func (ol *ObjectLock) GetRetention(bucket, key string) (*RetentionPolicy, bool) {
	ol.mu.RLock()
	defer ol.mu.RUnlock()

	policy, ok := ol.retention[fmt.Sprintf("%s/%s", bucket, key)]
	return policy, ok
}

// RemoveRetention removes retention policy
func (ol *ObjectLock) RemoveRetention(bucket, key, requestingUser string) error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	keyStr := fmt.Sprintf("%s/%s", bucket, key)
	policy, exists := ol.retention[keyStr]

	if !exists {
		return nil
	}

	// Check if it's COMPLIANCE mode
	if policy.Mode == "COMPLIANCE" {
		return fmt.Errorf("cannot remove COMPLIANCE retention")
	}

	// Check retention period
	if time.Now().Before(policy.RetainUntil) {
		return fmt.Errorf("object is under retention until %s", policy.RetainUntil)
	}

	delete(ol.retention, keyStr)
	return nil
}

// SetLegalHold sets legal hold on an object
func (ol *ObjectLock) SetLegalHold(bucket, key, status, createdBy string) error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	// Check if object lock is enabled
	if !ol.isLockEnabled(bucket) {
		return fmt.Errorf("object lock not enabled for bucket: %s", bucket)
	}

	keyStr := fmt.Sprintf("%s/%s", bucket, key)
	ol.legalHolds[keyStr] = &LegalHold{
		Bucket:    bucket,
		Key:       key,
		Status:    status,
		CreatedAt: time.Now(),
		CreatedBy: createdBy,
	}

	return nil
}

// GetLegalHold gets legal hold status
func (ol *ObjectLock) GetLegalHold(bucket, key string) (*LegalHold, bool) {
	ol.mu.RLock()
	defer ol.mu.RUnlock()

	hold, ok := ol.legalHolds[fmt.Sprintf("%s/%s", bucket, key)]
	return hold, ok
}

// DeleteLegalHold deletes legal hold
func (ol *ObjectLock) DeleteLegalHold(bucket, key string) error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	keyStr := fmt.Sprintf("%s/%s", bucket, key)
	delete(ol.legalHolds, keyStr)
	return nil
}

// CheckRetention checks if an object can be deleted based on retention
func (ol *ObjectLock) CheckRetention(bucket, key string) error {
	ol.mu.RLock()
	defer ol.mu.RUnlock()

	keyStr := fmt.Sprintf("%s/%s", bucket, key)

	// Check legal hold first
	if hold, ok := ol.legalHolds[keyStr]; ok {
		if hold.Status == "ON" {
			return fmt.Errorf("legal hold is active on object")
		}
	}

	// Check retention policy
	if policy, ok := ol.retention[keyStr]; ok {
		if time.Now().Before(policy.RetainUntil) {
			return fmt.Errorf("object is under retention until %s", policy.RetainUntil)
		}
	}

	return nil
}

// isLockEnabled checks if lock is enabled (internal)
func (ol *ObjectLock) isLockEnabled(bucket string) bool {
	lock, ok := ol.locks[bucket]
	return ok && lock.Enabled
}

// ListRetentions lists all retentions for a bucket
func (ol *ObjectLock) ListRetentions(bucket string) []*RetentionPolicy {
	ol.mu.RLock()
	defer ol.mu.RUnlock()

	var retentions []*RetentionPolicy
	for _, policy := range ol.retention {
		if policy.Bucket == bucket {
			retentions = append(retentions, policy)
		}
	}

	return retentions
}

// ListLegalHolds lists all legal holds for a bucket
func (ol *ObjectLock) ListLegalHolds(bucket string) []*LegalHold {
	ol.mu.RLock()
	defer ol.mu.RUnlock()

	var holds []*LegalHold
	for _, hold := range ol.legalHolds {
		if hold.Bucket == bucket {
			holds = append(holds, hold)
		}
	}

	return holds
}
