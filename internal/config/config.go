package config

import (
	"encoding/json"
	"flag"
	"os"
)

var filename string

func NewConfig() (*AppConfig, error) {
	flag.StringVar(&filename,
		"path-to-config",
		"config/app.conf",
		"path to configuration file",
	)

	flag.Parse()

	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config AppConfig

	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

type AppConfig struct {
	WorkerConfig       WorkerConfig  `json:"worker"`
	ServiceConfig      ServiceConfig `json:"service"`
	ChannelSize        int           `json:"queue_task_size"`
	CacheClearInterval int           `json:"cache_clear_interval"`
}

type ServiceConfig struct {
	IP   string `json:"IP"`
	PORT string `json:"PORT"`
}

type WorkerConfig struct {
	RemoteHTTPServer struct {
		IP   string `json:"IP"`
		PORT string `json:"PORT"`
	} `json:"remote_http_server"`
	Authentication struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	} `json:"authentication"`
	MaxWorkers         int `json:"max_workers"`
	MaxTimeForResponse int `json:"max_time_for_response"`
	TimeoutConnection  int `json:"timeout_connection"`
}
