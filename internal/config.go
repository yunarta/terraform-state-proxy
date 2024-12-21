package internal

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Bitbucket struct {
		Server string `toml:"server"`
	} `toml:"bitbucket"`

	Gitea struct {
		Server string `toml:"server"`
	} `toml:"gitea"`
}

func TestGetConfig() (*Config, error) {
	var config = Config{}
	var err error

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	execDir := filepath.Dir(execPath)

	configPaths := []string{
		filepath.Join(wd, "config.toml"),
		filepath.Join(execDir, "config.toml"),
		"/etc/tfstate/config.toml",
	}

	log.Println(configPaths)
	for _, path := range configPaths {
		log.Printf("Attempting to load config from: %s", path)
		if _, err = toml.DecodeFile(path, &config); err == nil {
			log.Printf("Successfully loaded config from: %s", path)
			return &config, nil
		}
	}

	return nil, err
}
