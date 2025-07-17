package main

import (
	"fmt"
	"log"
	"os"
)

// App은 애플리케이션의 모든 구성 요소와 상태를 관리하는 핵심 구조체입니다.
type App struct {
	config          *Config
	logger          *log.Logger
	device          DeviceCommunicator
	shutdownChan    chan struct{}
	filePath        string
	isAuthenticated bool // 인증 상태 저장
}

// NewApp은 새로운 App 인스턴스를 생성하고 초기화합니다.
func NewApp(config *Config, logger *log.Logger, filePath string) (*App, error) {
	var device DeviceCommunicator

	if config.Serial.Enabled {
		logger.Println("Serial communication enabled. Using RemoteDevice.")
		device = NewRemoteDevice(config.Serial.ServerURL, logger)
	} else {
		logger.Println("Serial communication disabled. Using MockDevice.")
		device = NewMockDevice(logger)
	}

	// 파일 존재 확인
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	return &App{
		config:       config,
		logger:       logger,
		device:       device,
		shutdownChan: make(chan struct{}),
		filePath:     filePath,
	}, nil
}
