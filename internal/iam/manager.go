package iam

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// User represents an IAM user
type User struct {
	ID            string            `json:"id"`
	TenantID      string            `json:"tenant_id"`
	Username      string            `json:"username"`
	Email         string            `json:"email"`
	Groups        []string          `json:"groups"`
	PolicyArns    []string          `json:"policy_arns"`
	InlinePolicy  *Policy           `json:"inline_policy,omitempty"`
	Status        string            `json:"status"` // active, inactive
	AccessKeys    []AccessKey       `json:"access_keys"`
	CreatedAt     time.Time         `json:"created_at"`
	LastActivity  time.Time         `json:"last_activity"`
}

// AccessKey represents an access key
type AccessKey struct {
	ID        string    `json:"id"`
	Secret    string    `json:"secret"`
	Status    string    `json:"status"` // active, inactive
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// Group represents an IAM group
type Group struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	PolicyArns  []string  `json:"policy_arns"`
	Members     []string  `json:"members"` // user IDs
	CreatedAt   time.Time `json:"created_at"`
}

// Policy represents an IAM policy
type Policy struct {
	ID          string      `json:"id"`
	TenantID    string      `json:"tenant_id"`
	Name        string      `json:"name"`
	Arn         string      `json:"arn"`
	Version     string      `json:"version"` // 2012-10-17
	Document    PolicyDoc   `json:"document"`
	IsAttached  bool        `json:"is_attached"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// PolicyDoc represents the policy document
type PolicyDoc struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

// Statement represents a policy statement
type Statement struct {
	Sid       string            `json:"Sid,omitempty"`
	Effect    string            `json:"Effect"` // Allow, Deny
	Actions   []string          `json:"Actions"`
	NotActions []string         `json:"NotActions,omitempty"`
	Resources []string          `json:"Resources"`
	NotResources []string       `json:"NotResources,omitempty"`
	Principals []Principal      `json:"Principals,omitempty"`
	NotPrincipals []string      `json:"NotPrincipals,omitempty"`
	Conditions map[string]map[string]interface{} `json:"Conditions,omitempty"`
}

// Principal represents a principal in a policy
type Principal struct {
	Type  string   `json:"type"` // AWS, Service, CanonicalUser, *
	Values []string `json:"values"`
}

// Role represents an IAM role
type Role struct {
	ID           string      `json:"id"`
	TenantID     string      `json:"tenant_id"`
	Name         string      `json:"name"`
	Arn          string      `json:"arn"`
	Path         string      `json:"path"`
	PolicyArns   []string    `json:"policy_arns"`
	AssumePolicy PolicyDoc   `json:"assume_policy"`
	CreatedAt    time.Time   `json:"created_at"`
}

// Manager manages IAM resources
type Manager struct {
	logger     *zap.Logger
	mu         sync.RWMutex
	users      map[string]*User
	groups     map[string]*Group
	policies   map[string]*Policy
	roles      map[string]*Role
}

// NewManager creates a new IAM manager
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		logger:   logger,
		users:    make(map[string]*User),
		groups:   make(map[string]*Group),
		policies: make(map[string]*Policy),
		roles:    make(map[string]*Role),
	}
}

// CreateUser creates a new user
func (m *Manager) CreateUser(tenantID, username, email string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user := &User{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		Username:     username,
		Email:        email,
		Groups:       []string{},
		PolicyArns:   []string{},
		Status:       "active",
		AccessKeys:   []AccessKey{},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	m.users[user.ID] = user
	m.logger.Info("User created",
		zap.String("id", user.ID),
		zap.String("username", username))

	return user, nil
}

// GetUser returns a user by ID
func (m *Manager) GetUser(userID string) (*User, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[userID]
	return u, ok
}

// GetUserByName returns a user by username
func (m *Manager) GetUserByName(tenantID, username string) (*User, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.TenantID == tenantID && u.Username == username {
			return u, true
		}
	}
	return nil, false
}

// ListUsers lists all users for a tenant
func (m *Manager) ListUsers(tenantID string) []*User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*User, 0)
	for _, u := range m.users {
		if u.TenantID == tenantID {
			result = append(result, u)
		}
	}
	return result
}

// DeleteUser deletes a user
func (m *Manager) DeleteUser(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.users[userID]; !ok {
		return fmt.Errorf("user not found: %s", userID)
	}

	delete(m.users, userID)
	m.logger.Info("User deleted", zap.String("id", userID))
	return nil
}

// CreateAccessKey creates an access key for a user
func (m *Manager) CreateAccessKey(userID string) (*AccessKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	key := AccessKey{
		ID:        "AKIA" + uuid.New().String()[:16],
		Secret:    uuid.New().String() + uuid.New().String(),
		Status:    "active",
		CreatedAt: time.Now(),
	}

	user.AccessKeys = append(user.AccessKeys, key)

	m.logger.Info("Access key created",
		zap.String("user_id", userID),
		zap.String("key_id", key.ID))

	return &key, nil
}

// CreateGroup creates a new group
func (m *Manager) CreateGroup(tenantID, name string) (*Group, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	group := &Group{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Name:      name,
		Members:   []string{},
		CreatedAt: time.Now(),
	}

	m.groups[group.ID] = group
	m.logger.Info("Group created",
		zap.String("id", group.ID),
		zap.String("name", name))

	return group, nil
}

// AddUserToGroup adds a user to a group
func (m *Manager) AddUserToGroup(userID, groupID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group, ok := m.groups[groupID]
	if !ok {
		return fmt.Errorf("group not found: %s", groupID)
	}

	user, ok := m.users[userID]
	if !ok {
		return fmt.Errorf("user not found: %s", userID)
	}

	// Check if user is already in group
	for _, member := range group.Members {
		if member == userID {
			return fmt.Errorf("user already in group")
		}
	}

	group.Members = append(group.Members, userID)
	user.Groups = append(user.Groups, groupID)

	m.logger.Info("User added to group",
		zap.String("user_id", userID),
		zap.String("group_id", groupID))

	return nil
}

// RemoveUserFromGroup removes a user from a group
func (m *Manager) RemoveUserFromGroup(userID, groupID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group, ok := m.groups[groupID]
	if !ok {
		return fmt.Errorf("group not found: %s", groupID)
	}

	// Remove from group
	newMembers := make([]string, 0)
	for _, member := range group.Members {
		if member != userID {
			newMembers = append(newMembers, member)
		}
	}
	group.Members = newMembers

	// Remove from user
	user, ok := m.users[userID]
	if ok {
		newGroups := make([]string, 0)
		for _, g := range user.Groups {
			if g != groupID {
				newGroups = append(newGroups, g)
			}
		}
		user.Groups = newGroups
	}

	m.logger.Info("User removed from group",
		zap.String("user_id", userID),
		zap.String("group_id", groupID))

	return nil
}

// CreatePolicy creates a new policy
func (m *Manager) CreatePolicy(tenantID, name string, doc PolicyDoc) (*Policy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	policy := &Policy{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		Name:       name,
		Arn:        fmt.Sprintf("arn:openendpoint:%s:%s:policy/%s", tenantID, "default", name),
		Version:    "2012-10-17",
		Document:   doc,
		IsAttached: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	m.policies[policy.ID] = policy
	m.logger.Info("Policy created",
		zap.String("id", policy.ID),
		zap.String("name", name))

	return policy, nil
}

// GetPolicy returns a policy by ID
func (m *Manager) GetPolicy(policyID string) (*Policy, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.policies[policyID]
	return p, ok
}

// GetPolicyByArn returns a policy by ARN
func (m *Manager) GetPolicyByArn(arn string) (*Policy, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.policies {
		if p.Arn == arn {
			return p, true
		}
	}
	return nil, false
}

// AttachPolicy attaches a policy to a user or group
func (m *Manager) AttachPolicy(policyID, entityID, entityType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	policy, ok := m.policies[policyID]
	if !ok {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	policy.IsAttached = true

	switch entityType {
	case "user":
		user, ok := m.users[entityID]
		if !ok {
			return fmt.Errorf("user not found: %s", entityID)
		}
		user.PolicyArns = append(user.PolicyArns, policy.Arn)
	case "group":
		group, ok := m.groups[entityID]
		if !ok {
			return fmt.Errorf("group not found: %s", entityID)
		}
		group.PolicyArns = append(group.PolicyArns, policy.Arn)
	default:
		return fmt.Errorf("invalid entity type: %s", entityType)
	}

	m.logger.Info("Policy attached",
		zap.String("policy_id", policyID),
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID))

	return nil
}

// DetachPolicy detaches a policy from a user or group
func (m *Manager) DetachPolicy(policyID, entityID, entityType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch entityType {
	case "user":
		user, ok := m.users[entityID]
		if !ok {
			return fmt.Errorf("user not found: %s", entityID)
		}
		newArns := make([]string, 0)
		for _, arn := range user.PolicyArns {
			if arn != policyID {
				newArns = append(newArns, arn)
			}
		}
		user.PolicyArns = newArns
	case "group":
		group, ok := m.groups[entityID]
		if !ok {
			return fmt.Errorf("group not found: %s", entityID)
		}
		newArns := make([]string, 0)
		for _, arn := range group.PolicyArns {
			if arn != policyID {
				newArns = append(newArns, arn)
			}
		}
		group.PolicyArns = newArns
	default:
		return fmt.Errorf("invalid entity type: %s", entityType)
	}

	m.logger.Info("Policy detached",
		zap.String("policy_id", policyID),
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID))

	return nil
}

// EvaluatePolicy evaluates if an action is allowed
func (m *Manager) EvaluatePolicy(tenantID, userID, action, resource string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[userID]
	if !ok {
		return false, fmt.Errorf("user not found: %s", userID)
	}

	// Get all policies for the user
	var policyArns []string
	policyArns = append(policyArns, user.PolicyArns...)

	// Also get policies from groups
	for _, groupID := range user.Groups {
		if group, ok := m.groups[groupID]; ok {
			policyArns = append(policyArns, group.PolicyArns...)
		}
	}

	// Evaluate each policy
	for _, arn := range policyArns {
		for _, policy := range m.policies {
			if policy.Arn == arn {
				allowed := m.evaluateStatement(policy.Document.Statement, action, resource)
				if allowed {
					return true, nil
				}
			}
		}
	}

	// Check inline policy
	if user.InlinePolicy != nil {
		allowed := m.evaluateStatement(user.InlinePolicy.Document.Statement, action, resource)
		if allowed {
			return true, nil
		}
	}

	return false, nil
}

// evaluateStatement evaluates a policy statement
func (m *Manager) evaluateStatement(statements []Statement, action, resource string) bool {
	for _, stmt := range statements {
		// Check effect
		if stmt.Effect != "Allow" {
			continue
		}

		// Check action
		actionMatched := false
		for _, a := range stmt.Actions {
			if a == action || a == "*" {
				actionMatched = true
				break
			}
		}
		if !actionMatched {
			continue
		}

		// Check resource
		resourceMatched := false
		for _, r := range stmt.Resources {
			if r == resource || r == "*" {
				resourceMatched = true
				break
			}
		}
		if !resourceMatched {
			continue
		}

		// All conditions met
		return true
	}

	return false
}

// PolicyFromJSON creates a policy from JSON
func PolicyFromJSON(data []byte) (*Policy, error) {
	var policy Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}
	return &policy, nil
}
