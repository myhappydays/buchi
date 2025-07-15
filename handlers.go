package main

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// FileInfo는 파일의 메타데이터를 담는 구조체입니다.
type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

// handleRoot는 다운로드 페이지를 제공합니다.
func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	ok, err := a.device.ValidateToken(token)
	if err != nil {
		a.logger.Printf("ERROR: Token validation failed: %v", err)
		http.Error(w, "Token validation error", http.StatusInternalServerError)
		return
	}
	if !ok {
		a.logger.Printf("WARN: Invalid token received: %s", token)
		http.Error(w, "Forbidden: Invalid Token", http.StatusForbidden)
		return
	}

	// download_page.html 파일을 직접 서빙합니다.
	http.ServeFile(w, r, "download_page.html")
}

// handleFileInfo는 공유할 파일의 정보를 JSON으로 반환합니다.
func (a *App) handleFileInfo(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	ok, err := a.device.ValidateToken(token)
	if err != nil {
		a.logger.Printf("ERROR: Token validation failed: %v", err)
		http.Error(w, "Token validation error", http.StatusInternalServerError)
		return
	}
	if !ok {
		a.logger.Printf("WARN: Invalid token received for fileinfo: %s", token)
		http.Error(w, "Forbidden: Invalid Token", http.StatusForbidden)
		return
	}

	stat, err := os.Stat(a.filePath)
	if err != nil {
		a.logger.Printf("ERROR: Failed to get file info: %v", err)
		http.Error(w, "Failed to read file information", http.StatusInternalServerError)
		return
	}

	fileName := filepath.Base(a.filePath)
	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	info := FileInfo{
		Name: fileName,
		Size: stat.Size(),
		Type: mimeType,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		a.logger.Printf("ERROR: Failed to encode file info: %v", err)
	}
}

// handleDownload는 실제 파일 다운로드를 처리합니다.
func (a *App) handleDownload(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	ok, err := a.device.ValidateToken(token)
	if err != nil {
		a.logger.Printf("ERROR: Token validation failed: %v", err)
		http.Error(w, "Token validation error", http.StatusInternalServerError)
		return
	}
	if !ok {
		a.logger.Printf("WARN: Invalid token received for download: %s", token)
		http.Error(w, "Forbidden: Invalid Token", http.StatusForbidden)
		return
	}

	file, err := os.Open(a.filePath)
	if err != nil {
		a.logger.Printf("ERROR: Failed to open file: %v", err)
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	fileName := filepath.Base(a.filePath)

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	a.logger.Printf("Starting download: %s (%d bytes)", fileName, stat.Size())

	_, err = io.Copy(w, file)
	if err != nil {
		a.logger.Printf("ERROR: File transfer failed: %v", err)
		// 클라이언트 연결이 끊어지는 등 전송 실패 시, 여기서 추가적인 에러 응답을 보내기 어려울 수 있습니다.
		// 따라서 로그만 남기고 함수를 종료합니다.
		return
	}

	a.logger.Printf("Download completed successfully: %s", fileName)

	// 다운로드가 성공적으로 완료되면, 앱 종료 신호를 보냅니다.
	go func() {
		// 핸들러가 응답을 완전히 보낼 시간을 주기 위해 잠시 대기합니다.
		// time.Sleep(1 * time.Second)
		close(a.shutdownChan) // 채널을 닫아 종료 신호를 보냅니다.
	}()
}