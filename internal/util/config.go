package util

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		DbSource string `yaml:"db_sorce"`
		DbName   string `yaml:"db_name"`
	} `yaml:"database"`
	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
}

func LoadConfig(filePath string) (Config, error) {
	var config Config

	file, err := os.Open(filePath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return config, err
	}
	return config, nil
}
