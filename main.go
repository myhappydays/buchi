package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// getLocalIP는 로컬 머신의 비-루프백 IPv4 주소를 찾습니다.
func getLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}
	for _, i := range interfaces {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			if ip.To4() != nil {
				return ip.String()
			}
		}
	}
	return "localhost"
}

func (a *App) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleRoot)
	mux.HandleFunc("/api/fileinfo", a.handleFileInfo)
	mux.HandleFunc("/download", a.handleDownload)

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", a.config.HTTPPort),
		Handler: mux,
	}

	// 서버를 고루틴으로 실행
	go func() {
		a.logger.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatalf("Could not start server: %v", err)
		}
	}()

	// 로컬 IP 주소를 가져와 URL을 장치에 씁니다.
	ip := getLocalIP()
	url := fmt.Sprintf("http://%s:%d", ip, a.config.HTTPPort)
	a.logger.Printf("Download URL: %s", url)
	if err := a.device.WriteURL(url); err != nil {
		a.logger.Printf("WARN: Failed to write URL to device: %v", err)
	}

	// 종료 신호 대기
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-a.shutdownChan: // 다운로드 완료 신호
		a.logger.Println("Download completed. Shutting down server...")
	case <-stop: // 사용자의 인터럽트 신호 (Ctrl+C)
		a.logger.Println("Interrupt signal received. Shutting down server...")
	}

	// 정상 종료 (Graceful Shutdown)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		a.logger.Fatalf("Server shutdown failed: %v", err)
	}

	a.logger.Println("Server stopped gracefully.")
	return a.device.Close()
}

func main() {
	logger := log.New(os.Stdout, "BUCHI | ", log.LstdFlags)

	if len(os.Args) < 2 {
		logger.Fatal("Usage: buchi <file_path>")
	}
	filePath := os.Args[1]

	config, err := LoadConfig("config.json")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	app, err := NewApp(config, logger, filePath)
	if err != nil {
		logger.Fatalf("Failed to create application: %v", err)
	}

	if err := app.Run(); err != nil {
		logger.Fatalf("Application run failed: %v", err)
	}

	logger.Println("Application finished.")
}
