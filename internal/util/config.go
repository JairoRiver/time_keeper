package util

import (
	"github.com/joho/godotenv"
)

type Config struct {
	DBSOURCE      string
	ServerAddress string
}

func LoadConfig(filePath string) (Config, error) {
	var myEnv map[string]string
	var config Config

	myEnv, err := godotenv.Read(filePath)
	if err != nil {
		return config, err
	}

	config.DBSOURCE = myEnv["DB_SOURCE"]
	config.ServerAddress = myEnv["SERVER_ADDRESS"]
	return config, nil
}
