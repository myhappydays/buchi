import serial
import time

# 테스트할 COM 포트
TEST_PORT = 'COM3'
BAUD_RATE = 115200

def test_port():
    print(f"Testing port {TEST_PORT}...")
    try:
        with serial.Serial(TEST_PORT, BAUD_RATE, timeout=2) as ser:
            time.sleep(2) # Wait for device to be ready
            print(f"Writing 'BUCHI:WHO' to {TEST_PORT}")
            ser.write(b"BUCHI:WHO\r\n")
            
            response = ser.readline().decode('utf-8').strip()
            
            if response:
                print(f"Received response: '{response}'")
                if "BUCHI:OK?" in response:
                    print("SUCCESS: Device responded correctly.")
                else:
                    print("FAILURE: Device responded, but with unexpected message.")
            else:
                print("FAILURE: No response received from the device.")
                
    except serial.SerialException as e:
        print(f"ERROR: Could not open or read from port {TEST_PORT}: {e}")

if __name__ == "__main__":
    test_port()
