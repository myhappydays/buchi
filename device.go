package main

import (
	"log"
)

// DeviceCommunicator는 하드웨어 장치와의 통신을 위한 인터페이스입니다.
type DeviceCommunicator interface {
	ValidateToken(token string) (bool, error)
	WriteURL(url string) error
	Close() error
}

// MockDevice는 테스트를 위한 가짜 장치 구현입니다.
type MockDevice struct {
	logger *log.Logger
}

// NewMockDevice는 MockDevice의 새 인스턴스를 생성합니다.
func NewMockDevice(logger *log.Logger) *MockDevice {
	return &MockDevice{logger: logger}
}

// ValidateToken은 항상 true를 반환하여 토큰 검증을 시뮬레이션합니다.
func (m *MockDevice) ValidateToken(token string) (bool, error) {
	m.logger.Printf("MockDevice: Validating token '%s' -> SUCCESS", token)
	return true, nil
}

// WriteURL은 주어진 URL을 콘솔에 로그로 남깁니다.
func (m *MockDevice) WriteURL(url string) error {
	m.logger.Printf("MockDevice: Writing URL -> %s", url)
	return nil
}

// Close는 아무 작업도 수행하지 않습니다.
func (m *MockDevice) Close() error {
	m.logger.Println("MockDevice: Closed.")
	return nil
}