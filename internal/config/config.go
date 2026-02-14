package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	CLIName             string  `mapstructure:"cli_name"`
	OllamaHost          string  `mapstructure:"ollama_host"`
	ModelSource         string  `mapstructure:"model_source"`
	VRAMOverheadPercent float64 `mapstructure:"vram_overhead_percent"`
	GPUMemoryFraction   float64 `mapstructure:"gpu_memory_fraction"`
}

func DefaultConfig() Config {
	return Config{
		CLIName:             "putzeys-cli",
		OllamaHost:          "http://localhost:11434",
		ModelSource:         "ollamadb",
		VRAMOverheadPercent: 20.0,
		GPUMemoryFraction:   0.75,
	}
}

func Load() (Config, error) {
	cfg := DefaultConfig()

	viper.SetDefault("cli_name", cfg.CLIName)
	viper.SetDefault("ollama_host", cfg.OllamaHost)
	viper.SetDefault("model_source", cfg.ModelSource)
	viper.SetDefault("vram_overhead_percent", cfg.VRAMOverheadPercent)
	viper.SetDefault("gpu_memory_fraction", cfg.GPUMemoryFraction)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	configDir, err := os.UserConfigDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(configDir, "putzeys-cli"))
	}
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("PUTZEYS")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return cfg, err
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
