package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v2"
)

type Log struct {
	Level string `yaml:"level"`
}

type Kafka struct {
	URL       string `yaml:"url"`
	MailTopic string `yaml:"mail_topic"`
	Partition string `yaml:"partition"`
}

type Sender struct {
	appName  string `yaml:"app_name"`
	RPS      int `yaml:"rps"`
	BuffSize int `yaml:"buff_size"`
}

type Config struct {
	Log     Log    `yaml:"logging"`
	Kafka   Kafka  `yaml:"kafka_config"`
	Sender Sender `yaml:"limiter_config"`
}

var (
	defaultConfig = "config/config.yaml"
	once          sync.Once
	configG       *Config
)

func NewConfig() *Config {
	var cfgPath string
	once.Do(func() {
		if c := os.Getenv("CFG_PATH"); c != "" {
			cfgPath = c
		} else {
			log.Println("if you are running localy -> export CFG_PATH=config/config_debug.yaml")
			cfgPath = defaultConfig
		}
		file, err := os.Open(filepath.Clean(cfgPath))
		if err != nil {
			log.Fatalf("config file problem %v", err)
		}
		defer file.Close()
		decoder := yaml.NewDecoder(file)
		configG = &Config{}
		err = decoder.Decode(configG)
		if err != nil {
			log.Fatalf("config file problem %v", err)
		}
	})
	return configG
}
