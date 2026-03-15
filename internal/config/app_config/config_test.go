package app_config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	appconfig "github.com/WithSoull/in-memory-database/internal/config/app_config"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := appconfig.DefaultConfig()

	require.Equal(t, "in_memory", cfg.Engine.Type)
	require.Equal(t, "127.0.0.1:3223", cfg.Network.Address)
	require.Equal(t, 100, cfg.Network.MaxConnections)
	require.Equal(t, "4KB", cfg.Network.MaxMessageSize)
	require.Equal(t, "5m", cfg.Network.IdleTimeout)
	require.Equal(t, 4*1024, cfg.Network.MaxMessageSizeBytes)
	require.Equal(t, 5*time.Minute, cfg.Network.IdleTimeoutDuration)
	require.Equal(t, "info", cfg.Logging.Level)
	require.Equal(t, "app.log", cfg.Logging.Output)

	require.False(t, cfg.WAL.Enabled)
	require.Equal(t, "./data/wal", cfg.WAL.DataPath)
	require.Equal(t, "16MB", cfg.WAL.SegmentMaxSize)
	require.Equal(t, 16*1024*1024, cfg.WAL.SegmentMaxSizeBytes)
	require.Equal(t, 100, cfg.WAL.BatchMaxLen)
	require.Equal(t, "10ms", cfg.WAL.BatchTimeout)
	require.Equal(t, 10*time.Millisecond, cfg.WAL.BatchTimeoutDuration)
}

func TestLoadWithEmptyPathReturnsDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := appconfig.Load("")
	require.NoError(t, err)
	expected := appconfig.DefaultConfig()
	require.Equal(t, &expected, cfg)
}

func TestLoadNonExistentFileReturnsDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := appconfig.Load("/nonexistent/path/config.yaml")
	require.NoError(t, err)
	expected := appconfig.DefaultConfig()
	require.Equal(t, &expected, cfg)
}

func TestLoadFullConfig(t *testing.T) {
	t.Parallel()

	content := `
engine:
  type: "in_memory"
network:
  address: "127.0.0.1:4000"
  max_connections: 25
  max_message_size: "8KB"
  idle_timeout: "45s"
logging:
  level: "debug"
  output: "/tmp/db.log"
`
	path := writeConfigFile(t, content)

	cfg, err := appconfig.Load(path)
	require.NoError(t, err)

	require.Equal(t, "in_memory", cfg.Engine.Type)
	require.Equal(t, "127.0.0.1:4000", cfg.Network.Address)
	require.Equal(t, 25, cfg.Network.MaxConnections)
	require.Equal(t, "8KB", cfg.Network.MaxMessageSize)
	require.Equal(t, "45s", cfg.Network.IdleTimeout)
	require.Equal(t, 8*1024, cfg.Network.MaxMessageSizeBytes)
	require.Equal(t, 45*time.Second, cfg.Network.IdleTimeoutDuration)
	require.Equal(t, "debug", cfg.Logging.Level)
	require.Equal(t, "/tmp/db.log", cfg.Logging.Output)
}

func TestLoadPartialConfigAppliesDefaults(t *testing.T) {
	t.Parallel()

	content := `
network:
  address: "0.0.0.0:9999"
  max_connections: 0
logging:
  level: "warn"
`
	path := writeConfigFile(t, content)

	cfg, err := appconfig.Load(path)
	require.NoError(t, err)

	require.Equal(t, "in_memory", cfg.Engine.Type)
	require.Equal(t, "0.0.0.0:9999", cfg.Network.Address)
	require.Equal(t, 100, cfg.Network.MaxConnections)
	require.Equal(t, "4KB", cfg.Network.MaxMessageSize)
	require.Equal(t, "5m", cfg.Network.IdleTimeout)
	require.Equal(t, 4*1024, cfg.Network.MaxMessageSizeBytes)
	require.Equal(t, 5*time.Minute, cfg.Network.IdleTimeoutDuration)
	require.Equal(t, "warn", cfg.Logging.Level)
	require.Equal(t, "app.log", cfg.Logging.Output)
}

func TestLoadFullWALConfig(t *testing.T) {
	t.Parallel()

	content := `
wal:
  enabled: true
  data_directory: "/tmp/wal"
  segment_max_size: "32MB"
  batch_max_size: 200
  batch_timeout: "20ms"
`
	path := writeConfigFile(t, content)

	cfg, err := appconfig.Load(path)
	require.NoError(t, err)

	require.True(t, cfg.WAL.Enabled)
	require.Equal(t, "/tmp/wal", cfg.WAL.DataPath)
	require.Equal(t, "32MB", cfg.WAL.SegmentMaxSize)
	require.Equal(t, 32*1024*1024, cfg.WAL.SegmentMaxSizeBytes)
	require.Equal(t, 200, cfg.WAL.BatchMaxLen)
	require.Equal(t, "20ms", cfg.WAL.BatchTimeout)
	require.Equal(t, 20*time.Millisecond, cfg.WAL.BatchTimeoutDuration)
}

func TestLoadPartialWALConfigAppliesDefaults(t *testing.T) {
	t.Parallel()

	content := `
wal:
  enabled: true
`
	path := writeConfigFile(t, content)

	cfg, err := appconfig.Load(path)
	require.NoError(t, err)

	require.True(t, cfg.WAL.Enabled)
	require.Equal(t, "./data/wal", cfg.WAL.DataPath)
	require.Equal(t, "16MB", cfg.WAL.SegmentMaxSize)
	require.Equal(t, 16*1024*1024, cfg.WAL.SegmentMaxSizeBytes)
	require.Equal(t, 100, cfg.WAL.BatchMaxLen)
	require.Equal(t, "10ms", cfg.WAL.BatchTimeout)
	require.Equal(t, 10*time.Millisecond, cfg.WAL.BatchTimeoutDuration)
}

func TestLoadInvalidWALSegmentMaxSize(t *testing.T) {
	t.Parallel()

	content := `
wal:
  segment_max_size: "4XB"
`
	path := writeConfigFile(t, content)
	_, err := appconfig.Load(path)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid WAL.segment_max_size")
}

func TestLoadInvalidWALBatchTimeout(t *testing.T) {
	t.Parallel()

	content := `
wal:
  batch_timeout: "not-a-duration"
`
	path := writeConfigFile(t, content)
	_, err := appconfig.Load(path)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid WAL.batch_timeout")
}

func TestLoadInvalidYAML(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "network: [")
	_, err := appconfig.Load(path)
	require.Error(t, err)
	require.ErrorContains(t, err, "parse config yaml")
}

func TestLoadInvalidMessageSize(t *testing.T) {
	t.Parallel()

	content := `
network:
  max_message_size: "4XB"
`
	path := writeConfigFile(t, content)
	_, err := appconfig.Load(path)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid network.max_message_size")
}

func TestLoadInvalidIdleTimeout(t *testing.T) {
	t.Parallel()

	content := `
network:
  idle_timeout: "not-a-duration"
`
	path := writeConfigFile(t, content)
	_, err := appconfig.Load(path)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid network.idle_timeout")
}

func TestParseMessageSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "bytes without unit",
			input: "512",
			want:  512,
		},
		{
			name:  "kilobytes",
			input: "4KB",
			want:  4 * 1024,
		},
		{
			name:  "megabytes lowercase",
			input: "2mb",
			want:  2 * 1024 * 1024,
		},
		{
			name:  "gigabytes with space",
			input: "1 GB",
			want:  1 * 1024 * 1024 * 1024,
		},
		{
			name:    "empty value",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid unit",
			input:   "4XB",
			wantErr: true,
		},
		{
			name:    "zero size",
			input:   "0KB",
			wantErr: true,
		},
		{
			name:    "missing numeric part",
			input:   "KB",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := appconfig.ParseMessageSize(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}
