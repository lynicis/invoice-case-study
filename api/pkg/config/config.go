package config

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	CorsOrigins string `koanf:"corsOrigins"`
	ServerPort  string `koanf:"serverPort"`
	Postgresql  struct {
		Host     string `koanf:"host"`
		Port     string `koanf:"port"`
		Username string `koanf:"username"`
		Password string `koanf:"password"`
		Database string `koanf:"database"`
	} `koanf:"postgresql"`
}

func Read() *Config {
	_, currentFile, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(currentFile), "../..")

	koanfInstance := koanf.New(".")
	configPath := file.Provider(filepath.Join(rootDir, "config/config.json"))
	if err := koanfInstance.Load(configPath, json.Parser()); err != nil {
		panic(fmt.Sprintf("error occurred while reading config: %s", err))
	}

	var config Config
	if err := koanfInstance.Unmarshal("", &config); err != nil {
		panic(fmt.Sprintf("error occurred while unmarshalling config: %s", err))
	}

	return &config
}
