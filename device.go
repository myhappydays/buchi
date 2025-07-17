package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

// RemoteDevice는 Python 시리얼 서버와 통신하는 실제 장치 구현입니다.
type RemoteDevice struct {
	serverURL string
	logger    *log.Logger
	client    *http.Client
}

// NewRemoteDevice는 RemoteDevice의 새 인스턴스를 생성합니다.
func NewRemoteDevice(serverURL string, logger *log.Logger) *RemoteDevice {
	return &RemoteDevice{
		serverURL: serverURL,
		logger:    logger,
		client:    &http.Client{},
	}
}

// ValidateToken은 Python 서버에 토큰 검증을 요청합니다.
func (r *RemoteDevice) ValidateToken(token string) (bool, error) {
	endpoint := r.serverURL + "/validate-token"
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return false, fmt.Errorf("invalid server url: %w", err)
	}

	q := reqURL.Query()
	q.Set("token", token)
	reqURL.RawQuery = q.Encode()

	r.logger.Printf("RemoteDevice: Validating token '%s' -> %s", token, reqURL.String())

	resp, err := r.client.Get(reqURL.String())
	if err != nil {
		return false, fmt.Errorf("failed to request token validation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusForbidden {
		return false, nil // 서버가 유효하지 않은 토큰으로 응답
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("server returned an error: %s - %s", resp.Status, string(body))
}

// WriteURL은 Python 서버에 URL 쓰기를 요청합니다.
func (r *RemoteDevice) WriteURL(urlString string) error {
	endpoint := r.serverURL + "/write-url"
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid server url: %w", err)
	}

	q := reqURL.Query()
	q.Set("url", urlString)
	reqURL.RawQuery = q.Encode()

	r.logger.Printf("RemoteDevice: Writing URL -> %s", reqURL.String())

	resp, err := r.client.Get(reqURL.String())
	if err != nil {
		return fmt.Errorf("failed to request url writing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned an error: %s - %s", resp.Status, string(body))
	}

	// 응답 본문 확인 (선택 사항)
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		r.logger.Printf("RemoteDevice: could not decode server response: %v", err)
	}
	r.logger.Printf("RemoteDevice: Server response -> %+v", result)

	return nil
}

// Close는 아무 작업도 수행하지 않습니다.
func (r *RemoteDevice) Close() error {
	r.logger.Println("RemoteDevice: Closed.")
	return nil
}
