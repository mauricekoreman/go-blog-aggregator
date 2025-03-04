package config

import (
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory")
		return "", err
	}

	configFilePath := fmt.Sprintf("%s/%s", homeDir, configFileName)

	return configFilePath, nil
}
