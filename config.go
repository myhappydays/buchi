package main

import (
	"encoding/json"
	"os"
)

// Config는 애플리케이션의 모든 설정을 담는 구조체입니다.
type Config struct {
	HTTPPort int `json:"http_port"`
	Serial   struct {
		BaudRate int  `json:"baud_rate"`
		Enabled  bool `json:"enabled"`
	} `json:"serial"`
}

// LoadConfig는 config.json 파일에서 설정을 읽어 Config 객체를 반환합니다.
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
