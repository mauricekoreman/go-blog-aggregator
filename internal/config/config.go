package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	DB_URL   string `json:"db_url"`
	Username string `json:"current_user_name"`
}

func Read() Config {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		fmt.Println("Error getting config file path")
		return Config{}
	}

	jsonFile, err := os.Open(configFilePath)
	if err != nil {
		fmt.Println("Error opening config file")
		return Config{}
	}

	defer jsonFile.Close()

	configData, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("Error reading config file")
		return Config{}
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Println("Error unmarshalling config data")
		return Config{}
	}

	return config
}

func (c *Config) SetUser(username string) error {
	c.Username = username

	updatedData, err := json.Marshal(c)
	if err != nil {
		fmt.Println("Error marshalling config data")
		return err
	}

	fileName, err := getConfigFilePath()
	if err != nil {
		fmt.Println("Error getting config file path")
		return err
	}

	err = os.WriteFile(fileName, updatedData, 0644)
	if err != nil {
		fmt.Println("Error writing to config file")
		return err
	}

	return nil
}
