package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// DbConfig is a struct used to capture configuration values from the yml file provided
type DbConfig struct {
	HOST     string `yaml:"host"`
	PORT     int64  `yaml:"port"`
	USER     string `yaml:"user"`
	PASSWORD string `yaml:"password"`
	DBNAME   string `yaml:"dbname"`
	SSLMODE  string `yaml:"sslmode"`
}

// GetDbConf returns the config struct used to initialise the postgres connection
func GetDbConf() *DbConfig {
	var c *DbConfig

	pwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Could not get root")
	}

	env := os.Getenv("ENV")

	if env == "" {
		env = "development"
	}

	yamlFilePath := filepath.Join(pwd, fmt.Sprintf("/config/%s.yml", env))
	yamlFile, err := ioutil.ReadFile(yamlFilePath)

	if err != nil {
		log.Fatalf("env.yml not provided")
		log.Panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &c)

	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
