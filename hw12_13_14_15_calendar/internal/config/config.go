package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Logger  LoggerConf    `json:"logger"`
	Storage StorageConfig `json:"storage"`
	Server  ServerConf    `json:"server"`
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
	HTTP Connection `json:"http"`
	Grpc Connection `json:"grpc"`
}

type Connection struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type QueueConf struct {
	DSN      string `json:"dsn"`
	Exchange string `json:"exchange"`
	Queue    string `json:"queue"`
}

type RemindsConf struct {
	RawRemindPeriod string `json:"remind_period"` //nolint:tagliatelle
	RawClearPeriod  string `json:"clear_period"`  //nolint:tagliatelle
	RemindPeriod    time.Duration
	ClearPeriod     time.Duration
}

type SchedulerConfig struct {
	Logger  LoggerConf    `json:"logger"`
	Storage StorageConfig `json:"storage"`
	Remind  RemindsConf   `json:"remind"`
	Queue   QueueConf     `json:"queue"`
}

type StorageConfig struct {
	MemoryStorage bool   `json:"memoryStorage"`
	Database      DBConf `json:"database"`
}

type SenderConfig struct {
	Logger LoggerConf `json:"logger"`
	Queue  QueueConf  `json:"queue"`
}

func NewConfig(confPath string) (Config, error) {
	config := Config{}
	rawCfg, err := getFileContent(confPath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(rawCfg, &config)
	return config, err
}

func NewSchedulerConfig(confPath string) (SchedulerConfig, error) {
	config := SchedulerConfig{}
	rawCfg, err := getFileContent(confPath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(rawCfg, &config)
	if err != nil {
		return config, err
	}
	config.Remind.RemindPeriod, err = time.ParseDuration(config.Remind.RawRemindPeriod)
	if err != nil {
		return config, fmt.Errorf("invalid config: error parsing remind_period: %w", err)
	}
	config.Remind.ClearPeriod, err = time.ParseDuration(config.Remind.RawClearPeriod)
	if err != nil {
		return config, fmt.Errorf("invalid config: error parsing clear_period: %w", err)
	}
	return config, nil
}

func NewSenderConfig(confPath string) (SenderConfig, error) {
	config := SenderConfig{}
	rawCfg, err := getFileContent(confPath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(rawCfg, &config)
	return config, err
}

func getFileContent(path string) ([]byte, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return []byte{}, fmt.Errorf("could not open config file: %w", err)
	}
	path = filepath.Join(pwd, path)
	file, err := os.Open(path)
	if err != nil {
		return []byte{}, fmt.Errorf("could not open config file: %w", err)
	}
	rawCfg, err := io.ReadAll(file)
	if err != nil {
		return []byte{}, fmt.Errorf("could not parse config, %w", err)
	}
	return rawCfg, nil
}
