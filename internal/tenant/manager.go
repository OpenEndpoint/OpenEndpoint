package tenant

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Tenant represents a tenant in the system
type Tenant struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Region      string    `json:"region"`
	Status      string    `json:"status"` // active, suspended, deleted
	Quota       *Quota    `json:"quota"`
	Settings    TenantSettings `json:"settings"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Quota contains resource quotas for a tenant
type Quota struct {
	StorageBytes   int64   `json:"storage_bytes"`   // Total storage in bytes
	ObjectCount    int64   `json:"object_count"`    // Max objects
	BucketCount    int     `json:"bucket_count"`    // Max buckets
	BandwidthMbps  int     `json:"bandwidth_mbps"`  // Max bandwidth
	APIRequests    int64   `json:"api_requests"`    // Max API requests per day
}

// TenantSettings contains tenant-specific settings
type TenantSettings struct {
	EncryptionRequired bool     `json:"encryption_required"`
	PublicAccessBlocked bool     `json:"public_access_blocked"`
	DefaultRegion       string   `json:"default_region"`
	AllowedRegions      []string `json:"allowed_regions"`
}

// Usage contains current usage statistics
type Usage struct {
	StorageBytes   int64   `json:"storage_bytes"`
	ObjectCount    int64   `json:"object_count"`
	BucketCount    int     `json:"bucket_count"`
	APIRequests    int64   `json:"api_requests"`
}

// Manager manages tenants
type Manager struct {
	logger   *zap.Logger
	mu       sync.RWMutex
	tenants  map[string]*Tenant
	quotas   map[string]*Usage
}

// NewManager creates a new tenant manager
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		logger:  logger,
		tenants: make(map[string]*Tenant),
		quotas:  make(map[string]*Usage),
	}
}

// CreateTenant creates a new tenant
func (m *Manager) CreateTenant(name, region string, quota *Quota) (*Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate name
	for _, t := range m.tenants {
		if t.Name == name && t.Status != "deleted" {
			return nil, fmt.Errorf("tenant with name %s already exists", name)
		}
	}

	tenant := &Tenant{
		ID:        uuid.New().String(),
		Name:      name,
		Region:    region,
		Status:    "active",
		Quota:     quota,
		Settings:  TenantSettings{
			DefaultRegion: region,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.tenants[tenant.ID] = tenant
	m.quotas[tenant.ID] = &Usage{}

	m.logger.Info("Tenant created",
		zap.String("id", tenant.ID),
		zap.String("name", name))

	return tenant, nil
}

// GetTenant returns a tenant by ID
func (m *Manager) GetTenant(tenantID string) (*Tenant, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tenants[tenantID]
	return t, ok
}

// GetTenantByName returns a tenant by name
func (m *Manager) GetTenantByName(name string) (*Tenant, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, t := range m.tenants {
		if t.Name == name && t.Status != "deleted" {
			return t, true
		}
	}
	return nil, false
}

// ListTenants lists all tenants
func (m *Manager) ListTenants() []*Tenant {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Tenant, 0, len(m.tenants))
	for _, t := range m.tenants {
		if t.Status != "deleted" {
			result = append(result, t)
		}
	}
	return result
}

// UpdateTenant updates a tenant
func (m *Manager) UpdateTenant(tenantID string, updates *Tenant) (*Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, ok := m.tenants[tenantID]
	if !ok {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	if updates.Name != "" {
		tenant.Name = updates.Name
	}
	if updates.Region != "" {
		tenant.Region = updates.Region
	}
	if updates.Quota != nil {
		tenant.Quota = updates.Quota
	}
	if updates.Settings.DefaultRegion != "" {
		tenant.Settings.DefaultRegion = updates.Settings.DefaultRegion
	}

	tenant.UpdatedAt = time.Now()

	return tenant, nil
}

// SuspendTenant suspends a tenant
func (m *Manager) SuspendTenant(tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, ok := m.tenants[tenantID]
	if !ok {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	tenant.Status = "suspended"
	tenant.UpdatedAt = time.Now()

	m.logger.Info("Tenant suspended", zap.String("id", tenantID))
	return nil
}

// ActivateTenant activates a tenant
func (m *Manager) ActivateTenant(tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, ok := m.tenants[tenantID]
	if !ok {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	tenant.Status = "active"
	tenant.UpdatedAt = time.Now()

	m.logger.Info("Tenant activated", zap.String("id", tenantID))
	return nil
}

// DeleteTenant deletes a tenant
func (m *Manager) DeleteTenant(tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, ok := m.tenants[tenantID]
	if !ok {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	// Soft delete
	tenant.Status = "deleted"
	tenant.UpdatedAt = time.Now()

	m.logger.Info("Tenant deleted", zap.String("id", tenantID))
	return nil
}

// GetUsage returns usage for a tenant
func (m *Manager) GetUsage(tenantID string) (*Usage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	usage, ok := m.quotas[tenantID]
	if !ok {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	return usage, nil
}

// UpdateUsage updates usage for a tenant
func (m *Manager) UpdateUsage(tenantID string, usage *Usage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.tenants[tenantID]; !ok {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	m.quotas[tenantID] = usage
	return nil
}

// CheckQuota checks if a tenant is within quota
func (m *Manager) CheckQuota(tenantID string, requiredBytes int64) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, ok := m.tenants[tenantID]
	if !ok {
		return false, fmt.Errorf("tenant not found: %s", tenantID)
	}

	if tenant.Status != "active" {
		return false, fmt.Errorf("tenant is not active")
	}

	usage, ok := m.quotas[tenantID]
	if !ok {
		return true, nil // No quota set
	}

	if tenant.Quota != nil {
		if usage.StorageBytes+requiredBytes > tenant.Quota.StorageBytes {
			return false, fmt.Errorf("storage quota exceeded")
		}
	}

	return true, nil
}

// AddStorageUsage adds storage usage
func (m *Manager) AddStorageUsage(tenantID string, bytes int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if usage, ok := m.quotas[tenantID]; ok {
		usage.StorageBytes += bytes
	}
	return nil
}

// AddAPIRequest adds an API request to usage
func (m *Manager) AddAPIRequest(tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if usage, ok := m.quotas[tenantID]; ok {
		usage.APIRequests++
	}
	return nil
}

// AddObjectCount adds object count
func (m *Manager) AddObjectCount(tenantID string, count int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if usage, ok := m.quotas[tenantID]; ok {
		usage.ObjectCount += count
	}
	return nil
}
