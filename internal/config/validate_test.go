package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
				},
				Storage: StorageConfig{
					DataDir: t.TempDir(),
				},
				Auth: AuthConfig{
					SecretKey: "test-secret-key-123",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Server: ServerConfig{
					Port: 0,
				},
				Storage: StorageConfig{
					DataDir: t.TempDir(),
				},
				Auth: AuthConfig{
					SecretKey: "test-secret-key-123",
				},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Server: ServerConfig{
					Port: 70000,
				},
				Storage: StorageConfig{
					DataDir: t.TempDir(),
				},
				Auth: AuthConfig{
					SecretKey: "test-secret-key-123",
				},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "missing data directory",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Storage: StorageConfig{
					DataDir: "",
				},
				Auth: AuthConfig{
					SecretKey: "test-secret-key-123",
				},
			},
			wantErr: true,
			errMsg:  "storage data directory is required",
		},
		{
			name: "missing secret key",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Storage: StorageConfig{
					DataDir: t.TempDir(),
				},
				Auth: AuthConfig{
					SecretKey: "",
				},
			},
			wantErr: true,
			errMsg:  "auth secret key is required",
		},
		{
			name: "short secret key",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Storage: StorageConfig{
					DataDir: t.TempDir(),
				},
				Auth: AuthConfig{
					SecretKey: "short",
				},
			},
			wantErr: true,
			errMsg:  "auth secret key must be at least 8 characters",
		},
		{
			name: "cluster enabled without node ID",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Storage: StorageConfig{
					DataDir: t.TempDir(),
				},
				Auth: AuthConfig{
					SecretKey: "test-secret-key-123",
				},
				Cluster: ClusterConfig{
					Enabled: true,
					NodeID:  "",
				},
			},
			wantErr: true,
			errMsg:  "cluster node ID is required",
		},
		{
			name: "invalid replication factor - too low",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Storage: StorageConfig{
					DataDir: t.TempDir(),
				},
				Auth: AuthConfig{
					SecretKey: "test-secret-key-123",
				},
				Cluster: ClusterConfig{
					Enabled:          true,
					NodeID:           "node1",
					ReplicationFactor: 0,
				},
			},
			wantErr: true,
			errMsg:  "cluster replication factor must be between 1 and 7",
		},
		{
			name: "invalid replication factor - too high",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Storage: StorageConfig{
					DataDir: t.TempDir(),
				},
				Auth: AuthConfig{
					SecretKey: "test-secret-key-123",
				},
				Cluster: ClusterConfig{
					Enabled:          true,
					NodeID:           "node1",
					ReplicationFactor: 10,
				},
			},
			wantErr: true,
			errMsg:  "cluster replication factor must be between 1 and 7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					// Partial match is OK for some errors
					if !contains(err.Error(), tt.errMsg) {
						t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
					}
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLoadConfig(t *testing.T) {
	// Test loading a non-existent config
	v := viper.New()
	_, err := LoadConfig(v)
	if err == nil {
		t.Error("LoadConfig should fail for non-existent config")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Port != 8080 {
		t.Errorf("Default port = %d, want 8080", cfg.Server.Port)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Default host = %s, want 0.0.0.0", cfg.Server.Host)
	}

	if cfg.Storage.DataDir != "/var/lib/openendpoint/data" {
		t.Errorf("Default data dir = %s, want /var/lib/openendpoint/data", cfg.Storage.DataDir)
	}

	if cfg.Auth.SecretKey == "" {
		t.Error("Default secret key should not be empty")
	}
}

func TestConfigWithDefaults(t *testing.T) {
	cfg := DefaultConfig()

	// Validate default config
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Default config should be valid, got: %v", err)
	}
}

func TestIsWritable(t *testing.T) {
	// Test with a writable temp directory
	tmpDir := t.TempDir()
	err := isWritable(tmpDir)
	if err != nil {
		t.Errorf("isWritable() should succeed for writable directory, got: %v", err)
	}

	// Test with a non-existent directory that can be created
	newDir := filepath.Join(t.TempDir(), "subdir")
	err = isWritable(newDir)
	if err != nil {
		t.Errorf("isWritable() should succeed for new directory, got: %v", err)
	}

	// Cleanup
	os.RemoveAll(newDir)
}
