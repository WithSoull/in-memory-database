package app

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

const (
	// Engine defaults
	defaultEngineType = "in_memory"

	// Network defaults
	defaultNetworkAddress = "127.0.0.1:3223"
	defaultMaxConnections = 100
	defaultMaxMessageSize = "4KB"
	defaultIdleTimeout    = "5m"

	// Logging defaults
	defaultLoggingLevel  = "info"
	defaultLoggingOutput = "app.log"

	// WAL defaults
	defaultWALEnabled        = false
	defaultWALDataPath       = "./data/wal"
	defaultWALSegmentMaxSize = "16MB"
	defaultWALBatchMaxLen    = 100
	defaultWALBatchTimeout   = "10ms"

	kilobyte int64 = 1024
)

type Config struct {
	Engine  EngineConfig  `yaml:"engine"`
	Network NetworkConfig `yaml:"network"`
	Logging LoggingConfig `yaml:"logging"`
	WAL     WALConfig     `yaml:"wal"`
}

type EngineConfig struct {
	Type string `yaml:"type"`
}

type NetworkConfig struct {
	Address        string `yaml:"address"`
	MaxConnections int    `yaml:"max_connections"`
	MaxMessageSize string `yaml:"max_message_size"`
	IdleTimeout    string `yaml:"idle_timeout"`

	MaxMessageSizeBytes int           `yaml:"-"`
	IdleTimeoutDuration time.Duration `yaml:"-"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

type WALConfig struct {
	Enabled        bool   `yaml:"enabled"`
	DataPath       string `yaml:"data_directory"`
	SegmentMaxSize string `yaml:"segment_max_size"`
	BatchMaxLen    int    `yaml:"batch_max_size"`
	BatchTimeout   string `yaml:"batch_timeout"`

	SegmentMaxSizeBytes  int           `yaml:"-"`
	BatchTimeoutDuration time.Duration `yaml:"-"`
}

func DefaultConfig() Config {
	cfg := Config{
		Engine: EngineConfig{
			Type: defaultEngineType,
		},
		Network: NetworkConfig{
			Address:        defaultNetworkAddress,
			MaxConnections: defaultMaxConnections,
			MaxMessageSize: defaultMaxMessageSize,
			IdleTimeout:    defaultIdleTimeout,
		},
		Logging: LoggingConfig{
			Level:  defaultLoggingLevel,
			Output: defaultLoggingOutput,
		},
		WAL: WALConfig{
			Enabled:        defaultWALEnabled,
			DataPath:       defaultWALDataPath,
			SegmentMaxSize: defaultWALSegmentMaxSize,
			BatchMaxLen:    defaultWALBatchMaxLen,
			BatchTimeout:   defaultWALBatchTimeout,
		},
	}
	if err := cfg.parseDerivedValues(); err != nil {
		panic(fmt.Sprintf("invalid default config: %v", err))
	}
	return cfg
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()
	if strings.TrimSpace(path) == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	if len(data) == 0 {
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config yaml: %w", err)
	}

	cfg.applyDefaults()
	if err := cfg.parseDerivedValues(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func ParseMessageSize(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, errors.New("empty size")
	}

	firstUnitCharIdx := -1
	for i, r := range value {
		if !unicode.IsDigit(r) {
			firstUnitCharIdx = i
			break
		}
	}

	var numberPart string
	var unitPart string

	if firstUnitCharIdx == -1 {
		numberPart = value
	} else {
		numberPart = value[:firstUnitCharIdx]
		unitPart = strings.TrimSpace(value[firstUnitCharIdx:])
	}

	if numberPart == "" {
		return 0, errors.New("size value is missing")
	}

	number, err := strconv.ParseInt(numberPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size number: %w", err)
	}
	if number <= 0 {
		return 0, errors.New("size must be greater than zero")
	}

	multiplier, err := sizeUnitMultiplier(unitPart)
	if err != nil {
		return 0, err
	}

	if number > math.MaxInt64/multiplier {
		return 0, errors.New("size overflow")
	}

	result := number * multiplier
	if result > int64(math.MaxInt) {
		return 0, errors.New("size exceeds int range")
	}

	return int(result), nil
}

func sizeUnitMultiplier(unit string) (int64, error) {
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "", "B":
		return 1, nil
	case "K", "KB", "KIB":
		return kilobyte, nil
	case "M", "MB", "MIB":
		return kilobyte * kilobyte, nil
	case "G", "GB", "GIB":
		return kilobyte * kilobyte * kilobyte, nil
	default:
		return 0, fmt.Errorf("unsupported size unit: %q", unit)
	}
}

func (c *Config) applyDefaults() {
	if strings.TrimSpace(c.Engine.Type) == "" {
		c.Engine.Type = defaultEngineType
	}

	if strings.TrimSpace(c.Network.Address) == "" {
		c.Network.Address = defaultNetworkAddress
	}

	if c.Network.MaxConnections <= 0 {
		c.Network.MaxConnections = defaultMaxConnections
	}

	if strings.TrimSpace(c.Network.MaxMessageSize) == "" {
		c.Network.MaxMessageSize = defaultMaxMessageSize
	}

	if strings.TrimSpace(c.Network.IdleTimeout) == "" {
		c.Network.IdleTimeout = defaultIdleTimeout
	}

	if strings.TrimSpace(c.Logging.Level) == "" {
		c.Logging.Level = defaultLoggingLevel
	}

	if strings.TrimSpace(c.Logging.Output) == "" {
		c.Logging.Output = defaultLoggingOutput
	}

	if strings.TrimSpace(c.WAL.DataPath) == "" {
		c.WAL.DataPath = defaultWALDataPath
	}

	if strings.TrimSpace(c.WAL.SegmentMaxSize) == "" {
		c.WAL.SegmentMaxSize = defaultWALSegmentMaxSize
	}

	if c.WAL.BatchMaxLen <= 0 {
		c.WAL.BatchMaxLen = defaultWALBatchMaxLen
	}

	if strings.TrimSpace(c.WAL.BatchTimeout) == "" {
		c.WAL.BatchTimeout = defaultWALBatchTimeout
	}
}

func (c *Config) parseDerivedValues() error {
	// Network
	messageSizeBytes, err := ParseMessageSize(c.Network.MaxMessageSize)
	if err != nil {
		return fmt.Errorf("invalid network.max_message_size: %w", err)
	}

	idleTimeout, err := time.ParseDuration(c.Network.IdleTimeout)
	if err != nil {
		return fmt.Errorf("invalid network.idle_timeout: %w", err)
	}
	if idleTimeout <= 0 {
		return errors.New("invalid network.idle_timeout: must be greater than zero")
	}

	c.Network.MaxMessageSizeBytes = messageSizeBytes
	c.Network.IdleTimeoutDuration = idleTimeout

	// WAL
	segmentMaxSizeBytes, err := ParseMessageSize(c.WAL.SegmentMaxSize)
	if err != nil {
		return fmt.Errorf("invalid WAL.segment_max_size: %w", err)
	}
	batchTimeoutDuration, err := time.ParseDuration(c.WAL.BatchTimeout)
	if err != nil {
		return fmt.Errorf("invalid WAL.batch_timeout: %w", err)
	}
	if batchTimeoutDuration <= 0 {
		return errors.New("invalid WAL.batch_timeout: must be greater than zero")
	}

	c.WAL.SegmentMaxSizeBytes = segmentMaxSizeBytes
	c.WAL.BatchTimeoutDuration = batchTimeoutDuration

	return nil
}
