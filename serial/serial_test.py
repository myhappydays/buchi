import time
import serial

# 시리얼 포트 설정 (아두이노가 연결된 COM 포트와 동일한지 다시 확인!)
# 예: 'COM3' 대신 'COM4', 'COM5' 등 실제 포트 번호로 변경 필요
ser = serial.Serial('COM3', 115200, timeout=1) 

print("시리얼 포트 열림:", ser.name)

time.sleep(2)  # 아두이노가 초기화될 시간을 주기 위해 잠시 대기

try:
    # 요청 전송 (아두이노가 \r\n을 모두 인식하도록)
    request_message = b'BUCHI:WRITE.URL?http://172.28.112.1:28244\r\n' 
    ser.write(request_message)
    print(f"요청 전송: {request_message.decode().strip()}")

    # 응답 수신
    # 아두이노가 \r\n으로 끝나는 응답을 줄 것이므로 readline()이 적합
    start_time = time.time()
    while True:
        if ser.in_waiting: # 수신할 데이터가 있는지 확인
            data = ser.readline().decode(errors='ignore').strip()
            if data: # 비어있지 않은 응답만 출력
                print("수신된 응답:", data)
                # 원하는 응답을 받았다면 루프 종료 (예: 'HELLO:ARDUINO')
                if "HELLO:ARDUINO" in data: 
                    print("성공적으로 응답을 받았습니다!")
                    break 
        
        # 타임아웃 처리: 너무 오래 기다리지 않도록
        if time.time() - start_time > 5: # 5초 이상 응답이 없으면 타임아웃
            print("응답 대기 시간 초과 (Timeout).")
            break
        
        time.sleep(0.01) # CPU 과부하 방지
        
except serial.SerialException as e:
    print(f"시리얼 통신 오류: {e}")
except Exception as e:
    print(f"예상치 못한 오류 발생: {e}")
finally:
    if ser.is_open:
        ser.close()
        print("시리얼 포트 닫힘.")