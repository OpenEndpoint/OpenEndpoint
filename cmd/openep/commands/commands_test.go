package commands

import (
	"strings"
	"testing"

	"github.com/openendpoint/openendpoint/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestParseS3Path(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantBucket  string
		wantKey     string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid path with key",
			path:       "s3://mybucket/mykey",
			wantBucket: "mybucket",
			wantKey:    "mykey",
			wantErr:    false,
		},
		{
			name:       "valid path without key",
			path:       "s3://mybucket",
			wantBucket: "mybucket",
			wantKey:    "",
			wantErr:    false,
		},
		{
			name:       "path with nested key",
			path:       "s3://mybucket/path/to/object",
			wantBucket: "mybucket",
			wantKey:    "path/to/object",
			wantErr:    false,
		},
		{
			name:        "empty bucket",
			path:        "s3://",
			wantErr:     true,
			errContains: "invalid S3 path",
		},
		{
			name:       "no s3 prefix",
			path:       "mybucket/mykey",
			wantBucket: "mybucket",
			wantKey:    "mykey",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := parseS3Path(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseS3Path(%q) expected error, got nil", tt.path)
				}
				return
			}
			if err != nil {
				t.Errorf("parseS3Path(%q) unexpected error: %v", tt.path, err)
				return
			}
			if bucket != tt.wantBucket {
				t.Errorf("parseS3Path(%q) bucket = %q, want %q", tt.path, bucket, tt.wantBucket)
			}
			if key != tt.wantKey {
				t.Errorf("parseS3Path(%q) key = %q, want %q", tt.path, key, tt.wantKey)
			}
		})
	}
}

func TestGetConfigValue(t *testing.T) {
	// Create a minimal config for testing
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 9000,
		},
		Storage: config.StorageConfig{
			DataDir:        "/data",
			StorageBackend: "flatfile",
		},
		LogLevel: "info",
	}

	tests := []struct {
		name     string
		key      string
		expected interface{}
	}{
		{"server.host", "server.host", "localhost"},
		{"server.port", "server.port", 9000},
		{"storage.data_dir", "storage.data_dir", "/data"},
		{"storage.backend", "storage.backend", "flatfile"},
		{"log_level", "log_level", "info"},
		{"unknown.key", "unknown.key", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getConfigValue(cfg, tt.key)
			if result != tt.expected {
				t.Errorf("getConfigValue(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestRootCmd(t *testing.T) {
	if RootCmd == nil {
		t.Fatal("RootCmd is nil")
	}
	if RootCmd.Use != "openep" {
		t.Errorf("RootCmd.Use = %q, want 'openep'", RootCmd.Use)
	}
}

func TestServerCmd(t *testing.T) {
	if ServerCmd == nil {
		t.Fatal("ServerCmd is nil")
	}
	if ServerCmd.Use != "server" {
		t.Errorf("ServerCmd.Use = %q, want 'server'", ServerCmd.Use)
	}
}

func TestVersionCmd(t *testing.T) {
	if VersionCmd == nil {
		t.Fatal("VersionCmd is nil")
	}
	if VersionCmd.Use != "version" {
		t.Errorf("VersionCmd.Use = %q, want 'version'", VersionCmd.Use)
	}
}

func TestBucketCmd(t *testing.T) {
	if BucketCmd == nil {
		t.Fatal("BucketCmd is nil")
	}
	if BucketCmd.Use != "bucket" {
		t.Errorf("BucketCmd.Use = %q, want 'bucket'", BucketCmd.Use)
	}
}

func TestBucketCreateCmd(t *testing.T) {
	if BucketCreateCmd == nil {
		t.Fatal("BucketCreateCmd is nil")
	}
	if BucketCreateCmd.Use != "create [bucket-name]" {
		t.Errorf("BucketCreateCmd.Use = %q, want 'create [bucket-name]'", BucketCreateCmd.Use)
	}
}

func TestBucketListCmd(t *testing.T) {
	if BucketListCmd == nil {
		t.Fatal("BucketListCmd is nil")
	}
	if BucketListCmd.Use != "ls" {
		t.Errorf("BucketListCmd.Use = %q, want 'ls'", BucketListCmd.Use)
	}
}

func TestBucketDeleteCmd(t *testing.T) {
	if BucketDeleteCmd == nil {
		t.Fatal("BucketDeleteCmd is nil")
	}
	if BucketDeleteCmd.Use != "rm [bucket-name]" {
		t.Errorf("BucketDeleteCmd.Use = %q, want 'rm [bucket-name]'", BucketDeleteCmd.Use)
	}
}

func TestBucketInfoCmd(t *testing.T) {
	if BucketInfoCmd == nil {
		t.Fatal("BucketInfoCmd is nil")
	}
	if BucketInfoCmd.Use != "info [bucket-name]" {
		t.Errorf("BucketInfoCmd.Use = %q, want 'info [bucket-name]'", BucketInfoCmd.Use)
	}
}

func TestObjectCmd(t *testing.T) {
	if ObjectCmd == nil {
		t.Fatal("ObjectCmd is nil")
	}
	if ObjectCmd.Use != "object" {
		t.Errorf("ObjectCmd.Use = %q, want 'object'", ObjectCmd.Use)
	}
}

func TestObjectPutCmd(t *testing.T) {
	if ObjectPutCmd == nil {
		t.Fatal("ObjectPutCmd is nil")
	}
	if ObjectPutCmd.Use != "put [source-file] s3://[bucket]/[key]" {
		t.Errorf("ObjectPutCmd.Use = %q, want 'put [source-file] s3://[bucket]/[key]'", ObjectPutCmd.Use)
	}
}

func TestObjectGetCmd(t *testing.T) {
	if ObjectGetCmd == nil {
		t.Fatal("ObjectGetCmd is nil")
	}
	if ObjectGetCmd.Use != "get s3://[bucket]/[key] [destination-file]" {
		t.Errorf("ObjectGetCmd.Use = %q, want 'get s3://[bucket]/[key] [destination-file]'", ObjectGetCmd.Use)
	}
}

func TestObjectListCmd(t *testing.T) {
	if ObjectListCmd == nil {
		t.Fatal("ObjectListCmd is nil")
	}
	if ObjectListCmd.Use != "ls s3://[bucket]" {
		t.Errorf("ObjectListCmd.Use = %q, want 'ls s3://[bucket]'", ObjectListCmd.Use)
	}
}

func TestObjectDeleteCmd(t *testing.T) {
	if ObjectDeleteCmd == nil {
		t.Fatal("ObjectDeleteCmd is nil")
	}
	if ObjectDeleteCmd.Use != "rm s3://[bucket]/[key]" {
		t.Errorf("ObjectDeleteCmd.Use = %q, want 'rm s3://[bucket]/[key]'", ObjectDeleteCmd.Use)
	}
}

func TestObjectCopyCmd(t *testing.T) {
	if ObjectCopyCmd == nil {
		t.Fatal("ObjectCopyCmd is nil")
	}
	if ObjectCopyCmd.Use != "cp s3://[source-bucket]/[source-key] s3://[dest-bucket]/[dest-key]" {
		t.Errorf("ObjectCopyCmd.Use = %q, want 'cp s3://[source-bucket]/[source-key] s3://[dest-bucket]/[dest-key]'", ObjectCopyCmd.Use)
	}
}

func TestConfigCmd(t *testing.T) {
	if ConfigCmd == nil {
		t.Fatal("ConfigCmd is nil")
	}
	if ConfigCmd.Use != "config" {
		t.Errorf("ConfigCmd.Use = %q, want 'config'", ConfigCmd.Use)
	}
}

func TestConfigGetCmd(t *testing.T) {
	if ConfigGetCmd == nil {
		t.Fatal("ConfigGetCmd is nil")
	}
	if ConfigGetCmd.Use != "get [key]" {
		t.Errorf("ConfigGetCmd.Use = %q, want 'get [key]'", ConfigGetCmd.Use)
	}
}

func TestConfigSetCmd(t *testing.T) {
	if ConfigSetCmd == nil {
		t.Fatal("ConfigSetCmd is nil")
	}
	if ConfigSetCmd.Use != "set [key] [value]" {
		t.Errorf("ConfigSetCmd.Use = %q, want 'set [key] [value]'", ConfigSetCmd.Use)
	}
}

func TestAdminCmd(t *testing.T) {
	if AdminCmd == nil {
		t.Fatal("AdminCmd is nil")
	}
	if AdminCmd.Use != "admin" {
		t.Errorf("AdminCmd.Use = %q, want 'admin'", AdminCmd.Use)
	}
}

func TestAdminInfoCmd(t *testing.T) {
	if AdminInfoCmd == nil {
		t.Fatal("AdminInfoCmd is nil")
	}
	if AdminInfoCmd.Use != "info" {
		t.Errorf("AdminInfoCmd.Use = %q, want 'info'", AdminInfoCmd.Use)
	}
}

func TestAdminStatsCmd(t *testing.T) {
	if AdminStatsCmd == nil {
		t.Fatal("AdminStatsCmd is nil")
	}
	if AdminStatsCmd.Use != "stats" {
		t.Errorf("AdminStatsCmd.Use = %q, want 'stats'", AdminStatsCmd.Use)
	}
}

func TestMonitorCmd(t *testing.T) {
	if MonitorCmd == nil {
		t.Fatal("MonitorCmd is nil")
	}
	if MonitorCmd.Use != "monitor" {
		t.Errorf("MonitorCmd.Use = %q, want 'monitor'", MonitorCmd.Use)
	}
}

func TestMonitorStatusCmd(t *testing.T) {
	if MonitorStatusCmd == nil {
		t.Fatal("MonitorStatusCmd is nil")
	}
	if MonitorStatusCmd.Use != "status" {
		t.Errorf("MonitorStatusCmd.Use = %q, want 'status'", MonitorStatusCmd.Use)
	}
}

func TestMonitorHealthCmd(t *testing.T) {
	if MonitorHealthCmd == nil {
		t.Fatal("MonitorHealthCmd is nil")
	}
	if MonitorHealthCmd.Use != "health" {
		t.Errorf("MonitorHealthCmd.Use = %q, want 'health'", MonitorHealthCmd.Use)
	}
}

func TestMonitorReadyCmd(t *testing.T) {
	if MonitorReadyCmd == nil {
		t.Fatal("MonitorReadyCmd is nil")
	}
	if MonitorReadyCmd.Use != "ready" {
		t.Errorf("MonitorReadyCmd.Use = %q, want 'ready'", MonitorReadyCmd.Use)
	}
}

func TestMonitorClusterCmd(t *testing.T) {
	if MonitorClusterCmd == nil {
		t.Fatal("MonitorClusterCmd is nil")
	}
	if MonitorClusterCmd.Use != "cluster" {
		t.Errorf("MonitorClusterCmd.Use = %q, want 'cluster'", MonitorClusterCmd.Use)
	}
}

func TestMonitorMetricsCmd(t *testing.T) {
	if MonitorMetricsCmd == nil {
		t.Fatal("MonitorMetricsCmd is nil")
	}
	if MonitorMetricsCmd.Use != "metrics" {
		t.Errorf("MonitorMetricsCmd.Use = %q, want 'metrics'", MonitorMetricsCmd.Use)
	}
}

func TestMonitorBucketsCmd(t *testing.T) {
	if MonitorBucketsCmd == nil {
		t.Fatal("MonitorBucketsCmd is nil")
	}
	if MonitorBucketsCmd.Use != "buckets" {
		t.Errorf("MonitorBucketsCmd.Use = %q, want 'buckets'", MonitorBucketsCmd.Use)
	}
}

func TestMonitorWatchCmd(t *testing.T) {
	if MonitorWatchCmd == nil {
		t.Fatal("MonitorWatchCmd is nil")
	}
	if MonitorWatchCmd.Use != "watch" {
		t.Errorf("MonitorWatchCmd.Use = %q, want 'watch'", MonitorWatchCmd.Use)
	}
}

func TestParseS3PathEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantBucket string
		wantKey    string
		wantErr    bool
	}{
		{"bucket with dots", "s3://my.bucket.name/key", "my.bucket.name", "key", false},
		{"bucket with hyphen", "s3://my-bucket/key", "my-bucket", "key", false},
		{"empty key", "s3://bucket/", "bucket", "", false},
		{"deeply nested", "s3://bucket/a/b/c/d/e", "bucket", "a/b/c/d/e", false},
		{"special chars in key", "s3://bucket/key-with_special.chars", "bucket", "key-with_special.chars", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := parseS3Path(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseS3Path(%q) expected error", tt.path)
				}
				return
			}
			if err != nil {
				t.Errorf("parseS3Path(%q) unexpected error: %v", tt.path, err)
				return
			}
			if bucket != tt.wantBucket {
				t.Errorf("parseS3Path(%q) bucket = %q, want %q", tt.path, bucket, tt.wantBucket)
			}
			if key != tt.wantKey {
				t.Errorf("parseS3Path(%q) key = %q, want %q", tt.path, key, tt.wantKey)
			}
		})
	}
}

func TestGetConfigValueAllCases(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Storage: config.StorageConfig{
			DataDir:        "/var/data",
			StorageBackend: "pebble",
		},
		LogLevel: "debug",
	}

	tests := []struct {
		key      string
		expected interface{}
	}{
		{"server.host", "0.0.0.0"},
		{"server.port", 8080},
		{"storage.data_dir", "/var/data"},
		{"storage.backend", "pebble"},
		{"log_level", "debug"},
		{"invalid.key", "unknown"},
		{"", "unknown"},
		{"server", "unknown"},
	}

	for _, tt := range tests {
		result := getConfigValue(cfg, tt.key)
		if result != tt.expected {
			t.Errorf("getConfigValue(%q) = %v, want %v", tt.key, result, tt.expected)
		}
	}
}

func TestBucketInfoStruct(t *testing.T) {
	info := BucketInfo{
		Name:        "test-bucket",
		ObjectCount: 42,
	}
	if info.Name != "test-bucket" {
		t.Errorf("BucketInfo.Name = %q, want 'test-bucket'", info.Name)
	}
	if info.ObjectCount != 42 {
		t.Errorf("BucketInfo.ObjectCount = %d, want 42", info.ObjectCount)
	}
}

func TestCommandDescriptions(t *testing.T) {
	commands := []struct {
		cmd  *cobra.Command
		name string
	}{
		{RootCmd, "RootCmd"},
		{ServerCmd, "ServerCmd"},
		{VersionCmd, "VersionCmd"},
		{BucketCmd, "BucketCmd"},
		{BucketCreateCmd, "BucketCreateCmd"},
		{BucketListCmd, "BucketListCmd"},
		{BucketDeleteCmd, "BucketDeleteCmd"},
		{BucketInfoCmd, "BucketInfoCmd"},
		{ObjectCmd, "ObjectCmd"},
		{ObjectPutCmd, "ObjectPutCmd"},
		{ObjectGetCmd, "ObjectGetCmd"},
		{ObjectListCmd, "ObjectListCmd"},
		{ObjectDeleteCmd, "ObjectDeleteCmd"},
		{ObjectCopyCmd, "ObjectCopyCmd"},
		{ConfigCmd, "ConfigCmd"},
		{ConfigGetCmd, "ConfigGetCmd"},
		{ConfigSetCmd, "ConfigSetCmd"},
		{AdminCmd, "AdminCmd"},
		{AdminInfoCmd, "AdminInfoCmd"},
		{AdminStatsCmd, "AdminStatsCmd"},
		{MonitorCmd, "MonitorCmd"},
		{MonitorStatusCmd, "MonitorStatusCmd"},
		{MonitorHealthCmd, "MonitorHealthCmd"},
		{MonitorReadyCmd, "MonitorReadyCmd"},
		{MonitorClusterCmd, "MonitorClusterCmd"},
		{MonitorMetricsCmd, "MonitorMetricsCmd"},
		{MonitorBucketsCmd, "MonitorBucketsCmd"},
		{MonitorWatchCmd, "MonitorWatchCmd"},
	}

	for _, tt := range commands {
		t.Run(tt.name+"_Short", func(t *testing.T) {
			if tt.cmd.Short == "" {
				t.Errorf("%s.Short should not be empty", tt.name)
			}
		})
		t.Run(tt.name+"_Long", func(t *testing.T) {
			// Long can be empty for simple commands
			_ = tt.cmd.Long
		})
	}
}

func TestCommandFlags(t *testing.T) {
	tests := []struct {
		cmd        *cobra.Command
		flagName   string
		exists     bool
		persistent bool
	}{
		{BucketDeleteCmd, "force", true, false},
		{ObjectDeleteCmd, "force", true, false},
		{ObjectListCmd, "recursive", true, false},
		{ObjectListCmd, "long", true, false},
		{MonitorWatchCmd, "interval", true, false},
		{RootCmd, "config", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.cmd.Use+"_"+tt.flagName, func(t *testing.T) {
			var flag *pflag.Flag
			if tt.persistent {
				flag = tt.cmd.PersistentFlags().Lookup(tt.flagName)
			} else {
				flag = tt.cmd.Flags().Lookup(tt.flagName)
			}
			if tt.exists && flag == nil {
				t.Errorf("flag %q should exist on %s", tt.flagName, tt.cmd.Use)
			}
			if !tt.exists && flag != nil {
				t.Errorf("flag %q should not exist on %s", tt.flagName, tt.cmd.Use)
			}
		})
	}
}

func TestGetServerURL(t *testing.T) {
	// Test with empty config path - should return default
	url := getServerURL()
	if url == "" {
		t.Error("getServerURL() should not return empty string")
	}
	// Should contain localhost or http
	if !strings.Contains(url, "http") {
		t.Errorf("getServerURL() = %q, should contain 'http'", url)
	}
}
