# 프로젝트 문서: BUCHI 파일 공유 시스템

## 1. 프로젝트 개요

이 프로젝트는 PC에 있는 파일을 하드웨어 장치(지금은 아두이노, 추후 esp32)를 통해 모바일 기기로 쉽고 안전하게 공유하기 위한 시스템입니다. 사용자는 PC에서 Go 애플리케이션을 실행하여 공유할 파일을 지정하고, 이 애플리케이션은 웹 서버를 구동하며 동시에 시리얼 통신을 통해 연결된 하드웨어 장치에 다운로드 URL을 전송합니다. 모바일 기기는 하드웨어 장치서 NFC를 통해 전송되는 URL을 통해 파일에 접근하고 다운로드할 수 있습니다.

**주요 목표:**
*   PC에서 모바일 기기로의 간편한 파일 전송.
*   하드웨어 장치를 통한 URL 전달 및 인증 토큰 관리.
*   안정적이고 효율적인 시리얼 통신 처리.

## 2. 프로젝트 구성

프로젝트는 크게 두 부분으로 나뉩니다: Go 기반의 HTTP 서버와 Python 기반의 시리얼 통신 서버.

```
Z:/Projects/2025_Buchi/
├───app.go              # Go 애플리케이션의 핵심 로직 및 초기화
├───config.go           # Go 애플리케이션 설정 로드
├───config.json         # Go 애플리케이션 설정 파일
├───device.go           # 장치 통신 인터페이스 및 Mock 구현
├───download_page.html  # 웹 다운로드 페이지 (클라이언트 UI)
├───go.mod              # Go 모듈 정의
├───go.sum              # Go 모듈 체크섬
├───handlers.go         # Go HTTP 요청 핸들러 (파일 정보, 다운로드 등)
├───main.go             # Go 애플리케이션 진입점 및 서버 실행
├───test.jpg            # 테스트용 이미지 파일
├───test.mp4            # 테스트용 비디오 파일
└───serial/             # 시리얼 통신 관련 Go 코드 (현재는 Python으로 대체)
    ├───serial_test.py  # Python 시리얼 시뮬레이터 (테스트용)
    └───serial.go       # Go 시리얼 통신 로직 (현재는 사용되지 않음)
└───serial_server/      # Python FastAPI 시리얼 서버
    ├───main_serial.py  # Python FastAPI 애플리케이션 진입점
    ├───serial_manager.py # Python 시리얼 통신 관리 클래스
    ├───requirements.txt  # Python 의존성 목록
    └───port_test.py    # Python 시리얼 포트 테스트 스크립트
```

### 2.1. Go 애플리케이션 (`.go` 파일들)

*   **역할:** 파일 공유를 위한 HTTP 서버를 구동하고, Python 시리얼 서버를 관리(실행/종료)합니다.
*   **주요 파일:**
    *   `main.go`: 애플리케이션의 진입점. `App` 인스턴스를 생성하고, HTTP 서버를 실행하며, Python 시리얼 서버를 시작/종료합니다.
    *   `app.go`: `App` 구조체를 정의하고 초기화합니다. 애플리케이션의 전반적인 상태와 의존성을 관리합니다.
    *   `handlers.go`: 웹 요청(`handleRoot`, `handleFileInfo`, `handleDownload`)을 처리합니다. 토큰 검증을 Python 시리얼 서버에 위임합니다.
    *   `config.go`, `config.json`: HTTP 포트 등 Go 애플리케이션의 설정을 관리합니다.
    *   `device.go`: `DeviceCommunicator` 인터페이스를 정의하며, 실제 장치 통신은 Python 서버로 위임됩니다.

### 2.2. Python 시리얼 서버 (`serial_server/` 디렉토리)

*   **역할:** 하드웨어 장치와의 실제 시리얼 통신을 담당합니다. Go 애플리케이션의 요청을 받아 장치에 URL을 쓰거나 토큰을 검증합니다.
*   **주요 파일:**
    *   `main_serial.py`: FastAPI 애플리케이션의 진입점. HTTP 엔드포인트(`write-url`, `validate-token`)를 제공하고 `SerialManager`를 사용하여 시리얼 통신을 수행합니다.
    *   `serial_manager.py`: `pyserial` 라이브러리를 사용하여 시리얼 포트를 관리합니다. 포트 검색, 열기, 데이터 읽기/쓰기, 토큰 검증, URL 쓰기 등의 로직을 포함합니다. `time.sleep()` 대신 타임아웃 기반의 논블로킹 읽기를 구현하여 효율성을 높였습니다.
    *   `requirements.txt`: Python 서버 실행에 필요한 라이브러리(`fastapi`, `uvicorn`, `pyserial`)를 정의합니다.

## 3. 작동 방식

1.  **애플리케이션 시작:**
    *   사용자가 Go 애플리케이션(`buchi.exe`)을 실행하고 공유할 파일 경로를 인자로 제공합니다.
    *   Go 애플리케이션은 `config.json`에서 설정을 로드합니다.
    *   Go 애플리케이션은 `os/exec`를 사용하여 Python FastAPI 시리얼 서버(`main_serial.py`)를 백그라운드에서 시작합니다.

2.  **Python 시리얼 서버 초기화:**
    *   Python 서버는 시작 시 `serial_manager.py`의 `find_and_open_port()` 함수를 호출하여 시스템의 모든 시리얼 포트를 스캔합니다.
    *   각 포트에 `BUCHI:WHO\r\n` 명령을 보내고, `BUCHI:OK?\r\n` 응답을 보내는 장치를 BUCHI 장치로 식별하여 연결합니다.
    *   이 과정에서 아두이노와 같은 장치의 리셋 시간을 고려하여 초기 통신 지연이 발생할 수 있습니다.

3.  **URL 전송 및 장치 연동:**
    *   Go 애플리케이션은 로컬 IP 주소와 설정된 HTTP 포트를 사용하여 다운로드 URL(예: `http://192.168.1.10:28244`)을 생성합니다.
    *   Go 애플리케이션은 이 URL을 Python 시리얼 서버의 `/write-url` 엔드포인트로 HTTP 요청을 통해 전송합니다.
    *   Python 시리얼 서버는 이 URL을 받아 시리얼 포트를 통해 하드웨어 장치에 `BUCHI:WRITE.URL?<url>\r\n` 명령으로 전송합니다.
    *   하드웨어 장치는 이 URL을 받아 자체적으로 표시(예: LCD, QR 코드)합니다.

4.  **모바일 기기 다운로드:**
    *   사용자는 모바일 기기로 하드웨어 장치에 표시된 URL(토큰 포함)에 접속합니다.
    *   모바일 기기의 웹 브라우저는 Go 애플리케이션의 HTTP 서버에 요청을 보냅니다.
    *   Go 애플리케이션의 핸들러(`handleRoot`, `handleFileInfo`, `handleDownload`)는 요청에 포함된 토큰을 추출합니다.
    *   Go 애플리케이션은 이 토큰을 Python 시리얼 서버의 `/validate-token` 엔드포인트로 HTTP 요청을 통해 전송합니다.
    *   Python 시리얼 서버는 토큰을 받아 시리얼 포트를 통해 하드웨어 장치에 `BUCHI:VALIDATE.TOKEN?<token>\r\n` 명령으로 전송하고, 장치로부터 `BUCHI:OK\r\n` 응답을 받으면 토큰이 유효하다고 판단합니다.
    *   토큰이 유효하면, Go 애플리케이션은 `download_page.html`을 제공하거나, 파일 정보(JSON)를 반환하거나, 실제 파일 다운로드를 시작합니다.

5.  **애플리케이션 종료:**
    *   파일 다운로드가 완료되거나, 사용자가 Go 애플리케이션을 종료하면, Go 애플리케이션은 실행 중인 Python 시리얼 서버 프로세스를 정상적으로 종료합니다.

## 4. 기술 스택

### 4.1. Go 애플리케이션

*   **언어:** Go
*   **웹 프레임워크:** Go 표준 라이브러리 (`net/http`)
*   **시리얼 통신 (간접):** Python 시리얼 서버와의 HTTP 통신을 통해 간접적으로 시리얼 통신을 수행.
*   **프로세스 관리:** `os/exec` 패키지를 사용하여 Python 서버 프로세스 실행 및 관리.

### 4.2. Python 시리얼 서버

*   **언어:** Python
*   **웹 프레임워크:** FastAPI (비동기 웹 프레임워크)
*   **ASGI 서버:** Uvicorn (FastAPI 애플리케이션을 실행하는 서버)
*   **시리얼 통신:** PySerial (시리얼 포트 통신 라이브러리)
*   **데이터 모델링:** Pydantic (FastAPI 요청 본문 유효성 검사 및 파싱)

### 4.3. 통신 프로토콜

*   **HTTP:** Go 애플리케이션과 Python 시리얼 서버 간의 통신, 그리고 Go 애플리케이션과 모바일 기기 간의 통신에 사용.
*   **시리얼 프로토콜 (Custom "BUCHI" Protocol):**
    *   `BUCHI:WHO\r\n`: 장치 식별 요청.
    *   `BUCHI:OK?\r\n`: 장치 식별 응답.
    *   `BUCHI:WRITE.URL?<url>\r\n`: 장치에 URL 쓰기 요청.
    *   `BUCHI:VALIDATE.TOKEN?<token>\r\n`: 장치에 토큰 유효성 검증 요청.
    *   `BUCHI:OK\r\n`: 장치로부터의 성공 응답.
    *   `BUCHI:INVALID\r\n`: 장치로부터의 유효하지 않음 응답.

```