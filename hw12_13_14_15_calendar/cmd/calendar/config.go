package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger        LoggerConf `json:"logger"`
	MemoryStorage bool       `json:"memoryStorage"`
	Database      DBConf     `json:"database"`
	Server        ServerConf `json:"server"`
}

type LoggerConf struct {
	Level       string `json:"logLevel"`
	Destination string `json:"destination"`
}

type DBConf struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"dbName"`
}

type ServerConf struct {
	Host string
	Port int
}

func NewConfig(confPath string) (Config, error) {
	config := Config{}
	pwd, err := os.Getwd()
	if err != nil {
		return config, fmt.Errorf("could not open config file: %w", err)
	}
	confPath = filepath.Join(pwd, confPath)
	file, err := os.Open(confPath)
	if err != nil {
		return config, fmt.Errorf("could not open config file: %w", err)
	}
	rawCfg, err := io.ReadAll(file)
	if err != nil {
		return config, fmt.Errorf("could not parse config, %w", err)
	}
	err = json.Unmarshal(rawCfg, &config)
	return config, err
}
