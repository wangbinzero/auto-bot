package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
)

type Env struct {
	Mode     string `yaml:"mode"`
	Node     string `yaml:"node"`
	Newrelic struct {
		AppName    string `yaml:"app_name"`
		LicenseKey string `yaml:"license_key"`
	} `yaml:"newrelic"`
}

var CurrentEnv Env

// initialize environment
func InitEnvironment() {

	//read the file from abslution path
	path, _ := filepath.Abs("config/env.yaml")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = yaml.Unmarshal(content, &CurrentEnv)
	if err != nil {
		log.Fatal(err)
	}
}
