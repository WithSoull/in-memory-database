package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	appconfig "github.com/WithSoull/in-memory-database/internal/config/app_config"
	zap_config "github.com/WithSoull/in-memory-database/internal/config/zap"
	"github.com/WithSoull/in-memory-database/internal/database"
	"github.com/WithSoull/in-memory-database/internal/database/compute/parser"
	"github.com/WithSoull/in-memory-database/internal/database/storage"
	inmemory "github.com/WithSoull/in-memory-database/internal/database/storage/engine/in_memory"
	tcpserver "github.com/WithSoull/in-memory-database/internal/network/tcp_server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultServerConfigPath = "server.config.yaml"

func main() {
	configPath := flag.String("config", defaultServerConfigPath, "path to config YAML (optional)")
	flag.Parse()

	if err := appconfig.InitConfig(*configPath); err != nil {
		panic(err)
	}

	logCfg := appconfig.Logging()
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logCfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	logger, err := zap_config.ZapConfig(level, logCfg.Output).Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	engineCfg := appconfig.Engine()
	if engineCfg.Type != "in_memory" {
		logger.Fatal("unsupported engine type", zap.String("type", engineCfg.Type))
	}

	compute := parser.NewParser(logger)

	engine, err := inmemory.NewEngine(logger)
	if err != nil {
		logger.Fatal("failed to create engine", zap.Error(err))
	}

	stor, err := storage.NewStrorage(engine, logger)
	if err != nil {
		logger.Fatal("failed to create storage", zap.Error(err))
	}

	db, err := database.NewDatabase(compute, stor, logger)
	if err != nil {
		logger.Fatal("failed to create database", zap.Error(err))
	}

	networkCfg := appconfig.Network()
	srv, err := tcpserver.NewServer(*networkCfg, logger)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	handler := func(ctx context.Context, req []byte) []byte {
		return []byte(db.HandleQuery(ctx, strings.TrimSpace(string(req))))
	}

	logger.Info("server started", zap.String("address", networkCfg.Address))
	srv.HandleQueries(ctx, handler)
	logger.Info("server stopped")
}
