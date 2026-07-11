package util

import (
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		DbSource string `yaml:"db_source"`
		DbName   string `yaml:"db_name"`
	} `yaml:"database"`
	Server struct {
		Address       string `yaml:"address"`
		SecureCookies bool   `yaml:"secure_cookies"`
		EnableSwagger bool   `yaml:"enable_swagger"`
	} `yaml:"server"`
}

// expandEnv replaces ${VAR} and ${VAR:-default} in s using environment variables.
func expandEnv(s string) string {
	return os.Expand(s, func(key string) string {
		if idx := strings.Index(key, ":-"); idx != -1 {
			if val := os.Getenv(key[:idx]); val != "" {
				return val
			}
			return key[idx+2:]
		}
		return os.Getenv(key)
	})
}

func LoadConfig(filePath string) (Config, error) {
	var config Config

	file, err := os.Open(filePath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return config, err
	}

	expanded := expandEnv(string(content))

	d := yaml.NewDecoder(strings.NewReader(expanded))
	if err := d.Decode(&config); err != nil {
		return config, err
	}
	return config, nil
}
