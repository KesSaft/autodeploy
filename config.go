package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type CustomCommand struct {
	command  string `json:"command"`
	force bool `json:"force"`
}

type Config struct {
	name string `json:"name"`
	key string `json:"key"`
	path string `json:"path"`
	externalPort int `json:"external_port"`
	internalPort int `json:"internal_port"`
	containerName string `json:"container_name"`
	githubToken string `json:"github_token"`
	dockerVolume bool `json:"docker_volume"`
	customVolume string `json:"custom_volume"`
	commands []CustomCommand  `json:"commands"`
}

func FindConfigWithSpecificValue(name string) ([]Config, error) {
	var foundConfig *Config

	err := filepath.Walk("./configs", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".json" {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			var config Config
			err = json.Unmarshal(data, &config)
			if err != nil {
				return err
			}

			if config.name == name {
				foundConfig = &config
				return errors.New("desired config found")
			}
		}
		return nil
	})

	if foundConfig != nil {
		return foundConfig, nil
	} else if err != nil && err.Error() == "desired config found" {
		return foundConfig, nil
	} else {
		return nil, errors.New("desired config not found")
	}
}