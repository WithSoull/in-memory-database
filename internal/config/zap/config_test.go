package zap_config_test

import (
	"os"
	"path/filepath"
	"testing"

	zap_config "github.com/WithSoull/in-memory-database/internal/config/zap"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestZapConfig_CreatesFileAtOutputPath(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "test.log")

	_, err := zap_config.ZapConfig(zapcore.InfoLevel, outputPath).Build()
	require.NoError(t, err)

	_, err = os.Stat(outputPath)
	require.NoError(t, err, "log file must be created at the given outputPath")
}

func TestZapConfig_LevelIsSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level zapcore.Level
	}{
		{name: "debug", level: zapcore.DebugLevel},
		{name: "info", level: zapcore.InfoLevel},
		{name: "warn", level: zapcore.WarnLevel},
		{name: "error", level: zapcore.ErrorLevel},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outputPath := filepath.Join(t.TempDir(), "test.log")
			cfg := zap_config.ZapConfig(tt.level, outputPath)

			require.Equal(t, tt.level, cfg.Level.Level())
		})
	}
}

func TestZapConfig_CanBuild(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "test.log")

	logger, err := zap_config.ZapConfig(zapcore.InfoLevel, outputPath).Build()
	require.NoError(t, err)
	require.NotNil(t, logger)
}
