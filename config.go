package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type Config struct {
	Host    string         `json:"host"`
	Port    int            `json:"port"`
	Storage StorageOptions `json:"storage"`
}

type StorageOptions struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DB       int    `json:"db"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoadConfig(filePath string) (config *Config, err error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.WithError(err).Error("Couldn't read config file")
		return nil, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		logrus.WithError(err).Error("Couldn't unmarshal configuration")
		return nil, err
	}

	return config, err
}
