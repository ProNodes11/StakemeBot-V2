package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Config struct {
	App  AppConfigData `yaml:"app"`
	Bot  BotConfigData `yaml:"bot"`
	Db   DbConfigData  `yaml:"database"`
	Chains []ChainConfigData `yaml:"chains"`
}

type AppConfigData struct {
	AppLogLevel string `yaml:"appLogLevel"`
}

type BotConfigData struct {
	BotDebug bool `yaml:"botDebug"`
	BotToken string `yaml:"botToken"`
}

type DbConfigData struct {
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type ChainConfigData struct {
	Name     string   `yaml:"name"`
	RPCUrls  []string `yaml:"rpcUrls"`
	APIUrls  []string `yaml:"apiUrls"`
}

func LoadConfig() (Config, error ) {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Error reading config file:", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("Error unmarshaling config:", err)
	}
	return config, nil
}

func (config *Config) GetAPIUrlsForChain(chainName string) string {
	for _, chain := range config.Chains {
		if chain.Name == chainName {
			return chain.APIUrls[0]
		}
	}
	return ""
}