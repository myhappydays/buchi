package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const passCookieName = "buchi_pass_id"

// generatePassID는 암호학적으로 안전한 랜덤 패스 ID를 생성합니다.
func generatePassID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// issueNewPass는 토큰을 검증하고, 성공 시 새로운 일회용 패스를 발급합니다.
func (a *App) issueNewPass(w http.ResponseWriter, r *http.Request) bool {
	token := r.URL.Query().Get("token")
	if token == "" {
		return false // 토큰이 없으면 발급 불가
	}

	ok, err := a.device.ValidateToken(token)
	if err != nil || !ok {
		return false
	}

	passID, err := generatePassID()
	if err != nil {
		a.logger.Printf("ERROR: Failed to generate pass ID: %v", err)
		return false
	}

	a.passMutex.Lock()
	a.downloadPasses[passID] = true
	a.passMutex.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     passCookieName,
		Value:    passID,
		Path:     "/",
		Expires:  time.Now().Add(5 * time.Minute), // 패스 유효기간 5분
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	a.logger.Printf("New download pass issued: %s", passID)
	return true
}

// hasValidPass는 요청에 유효한, 아직 사용되지 않은 패스가 있는지 확인합니다.
func (a *App) hasValidPass(r *http.Request) bool {
	cookie, err := r.Cookie(passCookieName)
	if err != nil {
		return false
	}

	a.passMutex.Lock()
	_, ok := a.downloadPasses[cookie.Value]
	a.passMutex.Unlock()

	return ok
}

// handleRoot는 다운로드 페이지를 제공하고, 필요 시 새로운 패스를 발급합니다.
func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	// 1. 유효한 패스가 이미 있는지 확인
	if a.hasValidPass(r) {
		http.ServeFile(w, r, "download_page.html")
		return
	}

	// 2. 패스가 없다면, URL 토큰으로 새 패스 발급 시도
	if !a.issueNewPass(w, r) {
		http.Error(w, "Forbidden: Invalid Token", http.StatusForbidden)
		return
	}

	// 3. 새 패스 발급에 성공했으므로, 페이지를 다시 로드하도록 유도
	// 쿠키가 설정된 후 페이지가 제대로 표시되려면 리디렉션이 가장 확실합니다.
	http.Redirect(w, r, "/", http.StatusFound)
}

// handleFileInfo는 유효한 패스가 있는 경우 파일 정보를 제공합니다.
func (a *App) handleFileInfo(w http.ResponseWriter, r *http.Request) {
	if !a.hasValidPass(r) {
		http.Error(w, "Forbidden: No valid download pass", http.StatusForbidden)
		return
	}

	stat, err := os.Stat(a.filePath)
	if err != nil {
		http.Error(w, "Failed to read file information", http.StatusInternalServerError)
		return
	}

	info := FileInfo{
		Name: filepath.Base(a.filePath),
		Size: stat.Size(),
		Type: mime.TypeByExtension(filepath.Ext(a.filePath)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleDownload는 패스를 소모하여 실제 파일 다운로드를 처리합니다.
func (a *App) handleDownload(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(passCookieName)
	if err != nil {
		http.Error(w, "Forbidden: No download pass cookie", http.StatusForbidden)
		return
	}
	passID := cookie.Value

	// 패스 유효성 검사 및 즉시 소모 (핵심 로직)
	a.passMutex.Lock()
	passExists, ok := a.downloadPasses[passID]
	if !ok || !passExists {
		a.passMutex.Unlock()
		http.Error(w, "Forbidden: Invalid or already used pass", http.StatusForbidden)
		return
	}
	// 패스를 즉시 삭제하여 재사용을 방지
	delete(a.downloadPasses, passID)
	a.passMutex.Unlock()

	a.logger.Printf("Consuming download pass: %s", passID)

	// 이하 파일 전송 로직
	file, err := os.Open(a.filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	fileName := filepath.Base(a.filePath)

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	_, err = io.Copy(w, file)
	if err != nil {
		a.logger.Printf("ERROR: File transfer failed: %v", err)
		return
	}

	a.logger.Printf("Download completed successfully for pass: %s", passID)

	go func() {
		close(a.shutdownChan)
	}()
}

// FileInfo는 파일의 메타데이터를 담는 구조체입니다.
type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}