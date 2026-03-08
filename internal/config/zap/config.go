package zap_config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ZapConfig(level zapcore.Level, outputPath string) zap.Config {
	return zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      []string{outputPath},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
	}
}
