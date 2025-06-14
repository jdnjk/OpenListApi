package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type ServiceConfig struct {
	Enable bool   `yaml:"enable"`
	UID    string `yaml:"uid"`
	Key    string `yaml:"key"`
}

type Config struct {
	Port     string        `yaml:"port"`
	Alipan   ServiceConfig `yaml:"alipan"`
	Baiduyun ServiceConfig `yaml:"baiduyun"`
	Pan123   ServiceConfig `yaml:"123pan"`
}

func LoadConfig() (*Config, error) {
	file, err := os.Open("config.yml")
	if os.IsNotExist(err) {
		defaultConfig := Config{
			Port: "8080",
			Alipan: ServiceConfig{
				Enable: false,
				UID:    "",
				Key:    "",
			},
			Baiduyun: ServiceConfig{
				Enable: false,
				UID:    "",
				Key:    "",
			},
			Pan123: ServiceConfig{
				Enable: false,
				UID:    "",
				Key:    "",
			},
		}
		data, _ := yaml.Marshal(&defaultConfig)
		os.WriteFile("config.yml", data, 0777)
		return nil, fmt.Errorf("请修改配置文件后操作。")
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
