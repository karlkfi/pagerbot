package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var Config config

func Load(configFilePath, envFilePath string) error {
	Config = config{}
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return fmt.Errorf("Config file not found: %s", configFilePath)
	}

	configBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("Failed to read config file: %s", err)
	}

	if envFilePath != "" {
		if _, err = os.Stat(envFilePath); os.IsNotExist(err) {
			return fmt.Errorf("Env file not found: %s", envFilePath)
		}

		// enf-file entries WILL OVERRIDE env variables that already exists
		err = godotenv.Load(envFilePath)
		if err != nil {
			return fmt.Errorf("Failed to load env file: %s", err)
		}
	}

	// expand env vars in yaml
	configString := os.ExpandEnv(string(configBytes))

	err = yaml.Unmarshal([]byte(configString), &Config)
	if err != nil {
		return fmt.Errorf("Failed to parse config file: %s", err)
	}

	return nil
}
