package main

import (
	"flag"
)

var (
	configPath string
)

func parseFlags() error {
	flag.StringVar(&configPath, "conf", "", "Путь к конфигурационному файлу")

	flag.Parse()

	if configPath == "" {
		configPath = "config/default.json"
	}
	return nil
}