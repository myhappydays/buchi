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

// isAuthorized는 요청이 유효한지 확인하고 세션 기반으로 인증을 처리합니다.
func (a *App) isAuthorized(token string) (bool, error) {
	if a.isAuthenticated {
		a.logger.Println("Session already authenticated.")
		return true, nil
	}

	a.logger.Println("Attempting to validate token with device...")
	ok, err := a.device.ValidateToken(token)
	if err != nil {
		return false, err
	}

	if ok {
		a.logger.Println("Token validation successful. Session is now authenticated.")
		a.isAuthenticated = true
	}

	return ok, nil
}

// handleRoot는 다운로드 페이지를 제공합니다.
func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	ok, err := a.isAuthorized(token)
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
	// 파일 정보는 인증 없이 접근을 허용할 수 있지만, 여기서는 일관성을 위해 인증을 유지합니다.
	token := r.URL.Query().Get("token")
	ok, err := a.isAuthorized(token)
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
	// 다운로드 요청 시에는 반드시 인증 상태를 확인해야 합니다.
	token := r.URL.Query().Get("token")
	ok, err := a.isAuthorized(token)
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
		return
	}

	a.logger.Printf("Download completed successfully: %s", fileName)

	go func() {
		close(a.shutdownChan)
	}()
}