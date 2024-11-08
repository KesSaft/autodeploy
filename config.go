package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type CustomCommand struct {
	Command string `json:"command"`
	Force   bool   `json:"force"`
}

type Config struct {
	Name              string          `json:"name"`
	Key               string          `json:"key"`
	Path              string          `json:"path"`
	ExternalPort      int             `json:"externalPort"`
	InternalPort      int             `json:"internalPort"`
	ContainerName     string          `json:"containerName"`
	GithubToken       string          `json:"githubToken"`
	DockerVolume      bool            `json:"dockerVolume"`
	CustomVolume      string          `json:"customVolume"`
	Branch            string          `json:"branch"`
	Seamless          bool            `json:"seamless"`
	ReadyForUpdateURL string          `json:"readyForUpdateWebhook"`
	Commands          []CustomCommand `json:"commands"`
}

func FindConfigWithSpecificValue(name string) (*Config, error) {
	var foundConfig *Config

	err := filepath.Walk("/projects/configs", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".json" {
			if info.IsDir() {
				return nil
			}

			rawJson, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			var config Config
			err = json.Unmarshal(rawJson, &config)

			if err != nil {
				return err
			}

			if config.Name == name {
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
