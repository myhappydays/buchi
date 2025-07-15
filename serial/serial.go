package serial

import (
	"fmt"
	"strings"
	"time"

	"go.bug.st/serial"
)

type Port = serial.Port

func GetPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get serial ports: %w", err)
	}

	var availablePorts []string
	for _, port := range ports {
		p, err := serial.Open(port, &serial.Mode{BaudRate: 115200})
		if err == nil {
			time.Sleep(5 * time.Second)
			request := []byte("BUCHI:WHO\r\n")
			fmt.Printf("Sending request to %s: %s\n", port, request)
			_, err = p.Write(request)
			if err == nil {
				time.Sleep(2 * time.Second) // 응답 대기 시간
				buf := make([]byte, 64)
				p.SetReadTimeout(0)
				n, _ := p.Read(buf)
				if n > 0 {
					data := buf[:n]
					payload := string(data)
					// "BUCHI:OK?"로 시작하는지 확인
					if len(payload) >= 9 && payload[:9] == "BUCHI:OK?" {
						availablePorts = append(availablePorts, port)
					} else {
						fmt.Printf("Unexpected response from %s: %s\n", port, payload)
					}
				} else {
					fmt.Printf("No data received from %s\n", port)
				}
			} else {
				fmt.Printf("Failed to write to port %s: %v\n", port, err)
			}
			p.Close()
		}
	}

	if len(availablePorts) == 0 {
		return nil, fmt.Errorf("no available serial ports found")
	}

	return availablePorts, nil
}

func OpenPort(portName string, baudRate int) (serial.Port, error) {
	mode := &serial.Mode{BaudRate: baudRate}
	p, err := serial.Open(portName, mode)

	if err != nil {
		return nil, fmt.Errorf("failed to open port %s: %w", portName, err)
	}
	fmt.Printf("Opened port: %s\n", portName)

	return p, nil
}

func WriteUrl(url string, p serial.Port) error {
	time.Sleep(5 * time.Second)

	request := []byte("BUCHI:WRITE.URL?" + url + "\r\n")
	fmt.Printf("Sending URL to port: %s\n", request)
	_, err := p.Write(request)
	if err != nil {
		return fmt.Errorf("failed to write URL to port: %w", err)
	}

	time.Sleep(2 * time.Second)
	buf := make([]byte, 64)
	p.SetReadTimeout(0)
	n, _ := p.Read(buf)

	if n > 0 {
		data := buf[:n]
		payload := string(data)
		fmt.Printf("Received response: %s\n", payload)
	} else {
		return fmt.Errorf("no data received from port")
	}
	return nil
}

func ValidateToken(token string, p serial.Port) (bool, error) {
	fmt.Println("validateToken 호출됨, token:", token)
	// --- 버퍼 비우기 ---
	buf := make([]byte, 64)
	for {
		p.SetReadTimeout(50)
		n, _ := p.Read(buf)
		if n == 0 {
			break
		}
	}
	// ------------------

	request := []byte("BUCHI:VALIDATE.TOKEN?" + token + "\r\n")
	_, err := p.Write(request)
	if err != nil {
		return false, fmt.Errorf("failed to write token validation: %w", err)
	}

	var response []byte
	start := time.Now()
	buf = make([]byte, 64)
	p.SetReadTimeout(200)
	for {
		n, _ := p.Read(buf)
		if n > 0 {
			response = append(response, buf[:n]...)
			if response[len(response)-1] == '\n' {
				break
			}
		}
		if time.Since(start) > 3*time.Second {
			break
		}
	}
	payload := string(response)
	fmt.Printf("Token validation response: %s\n", payload)
	if strings.Contains(payload, "BUCHI:OK") {
		return true, nil
	}
	return false, nil
}
