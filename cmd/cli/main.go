package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	zap_config "github.com/WithSoull/in-memory-database/internal/config/zap"
	"github.com/WithSoull/in-memory-database/internal/database"
	"github.com/WithSoull/in-memory-database/internal/database/compute/parser"
	"github.com/WithSoull/in-memory-database/internal/database/storage"
	inmemory "github.com/WithSoull/in-memory-database/internal/database/storage/engine/in_memory"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap_config.ZapConfig(zap.DebugLevel).Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	compute := parser.NewParser(logger)

	engine, err := inmemory.NewEngine(logger)
	if err != nil {
		logger.Fatal("failed to create enine", zap.Error(err))
	}

	storage, err := storage.NewStrorage(engine, logger)
	if err != nil {
		logger.Fatal("failed to create storage", zap.Error(err))
	}

	db, err := database.NewDatabase(compute, storage, logger)
	if err != nil {
		logger.Fatal("failed to create database", zap.Error(err))
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("In-memory DB (type EXIT to quit)")

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			logger.Error("failed to read line", zap.Error(err))
			break
		}

		line = strings.TrimSpace(line)
		if line == "EXIT" {
			fmt.Println("Bye!")
			break
		}

		ctx := context.Background()
		result := db.HandleQuery(ctx, line)
		fmt.Println(result)
	}
}
