package core

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type ProxyConfig struct {
	Apps []*AppConfig
}

type AppConfig struct {
	Name    string
	Ports   []int
	Targets []string
}

func ReadConfig(path string) (*ProxyConfig, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var config *ProxyConfig
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	return config, err
}
