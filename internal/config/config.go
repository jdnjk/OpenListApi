package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port         string `yaml:"port"`
	AliClientKey string `yaml:"ali_client_key"`
	AliClientUID string `yaml:"ali_client_uid"`
	Alipan       bool   `yaml:"alipan"`
}

func LoadConfig() (*Config, error) {
	file, err := os.Open("config.yml")
	if os.IsNotExist(err) {
		defaultConfig := Config{
			Port:         "8080",
			AliClientKey: "",
			AliClientUID: "",
			Alipan:       false,
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
