package app_config

var globalConfig *Config

func InitConfig(path string) error {
	var err error
	globalConfig, err = Load(path)
	if err != nil {
		return err
	}
	return nil
}

func GetConfig() *Config {
	return globalConfig
}

func Engine() *EngineConfig {
	if globalConfig != nil {
		return &globalConfig.Engine
	}
	return nil
}

func Network() *NetworkConfig {
	if globalConfig != nil {
		return &globalConfig.Network
	}
	return nil
}

func Logging() *LoggingConfig {
	if globalConfig != nil {
		return &globalConfig.Logging
	}
	return nil
}

func WAL() *WALConfig {
	if globalConfig != nil {
		return &globalConfig.WAL
	}
	return nil
}
