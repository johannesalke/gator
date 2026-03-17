package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}

func getConfigFilePath() string {
	homeDirPath, _ := os.UserHomeDir()
	configFilePath := filepath.Join(homeDirPath, configFileName)
	return configFilePath
}

func (cfg *Config) write() error {
	f, err := os.OpenFile(getConfigFilePath(), os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Error while opening config file: %s", err)
	}
	defer f.Close()

	fileEncoder := json.NewEncoder(f)
	err = fileEncoder.Encode(cfg)
	if err != nil {
		return fmt.Errorf("Error while encoding config json: %s", err)
	}
	return nil
}

func Read() *Config {

	f, err := os.Open(getConfigFilePath())
	if err != nil {
		fmt.Printf("Error while reading config file: %s", err)
		return nil
	}
	defer f.Close()

	fDecoder := json.NewDecoder(f)
	var cfg *Config

	err = fDecoder.Decode(&cfg)
	if err != nil {
		fmt.Printf("Error while reading file into config struct: %s", err)
		return nil
	}
	return cfg
}

func (cfg Config) SetUser(new_user_name string) error {
	cfg.CurrentUsername = new_user_name
	err := cfg.write()
	if err != nil {
		return fmt.Errorf("Error while reading file into config struct: %s", err)
	}
	return nil
}
