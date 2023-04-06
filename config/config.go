package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	MQTT struct {
		Host     string `yaml:"host"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		ClientId string `yaml:"clientId"`
	} `yaml:"mqtt"`
	RedisHost string `yaml:"redisHost"`
	Monitors  []struct {
		Name          string        `yaml:"name"`
		FriendlyName  string        `yaml:"friendlyName"`
		ReportTimeout time.Duration `yaml:"reportTimeout"`
		AssumeUp      bool          `yaml:"assumeUp"`
		MQTT          struct {
			Topic       string `yaml:"topic"`
			UpMessage   string `yaml:"upMessage"`
			DownMessage string `yaml:"downMessage"`
			Host        string `yaml:"host"`
			User        string `yaml:"user"`
			Password    string `yaml:"password"`
		}
	} `yaml:"monitors"`
}

var AppConfig Config

func init() {
	fp := os.Getenv("OHFUCK_CONFIG_FILE")
	if fp == "" {
		panic("OHFUCK_CONFIG_FILE is not set")
	}

	dat, err := os.ReadFile(fp)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(dat, &AppConfig)
	if err != nil {
		panic(err)
	}
}
