package zap_config

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ZapConfig(level zapcore.Level) zap.Config {
	_, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	return zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      []string{"app.log"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
	}
}
