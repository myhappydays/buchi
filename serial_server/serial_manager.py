import serial
import serial.tools.list_ports
import time

class SerialManager:
    def __init__(self, baud_rate=115200, timeout=1):
        self.port = None
        self.baud_rate = baud_rate
        self.timeout = timeout

    def find_and_open_port(self):
        ports = serial.tools.list_ports.comports()
        for port_info in ports:
            ser = None
            try:
                ser = serial.Serial(port_info.device, self.baud_rate, timeout=self.timeout)
                time.sleep(2) # Give the device some time to initialize
                ser.write(b"BUCHI:WHO\r\n")
                response = ser.readline().decode('utf-8').strip()
                if "BUCHI:OK?" in response:
                    print(f"Successfully connected to BUCHI device on port: {port_info.device}")
                    self.port = ser # Keep the port open
                    return True
                else:
                    ser.close() # Close if not the correct device
            except (OSError, serial.SerialException) as e:
                print(f"Could not open or write to port {port_info.device}: {e}")
                if ser and ser.is_open:
                    ser.close()
        
        print("Could not find any BUCHI device.")
        return False

    def write_and_read(self, command):
        if not self.port or not self.port.is_open:
            raise ConnectionError("Serial port is not open.")
        
        try:
            self.port.write(command.encode('utf-8'))
            response = self.port.readline().decode('utf-8').strip()
            print(f"[SerialManager] Received raw response: '{response}'") # Debug print
            return response
        except serial.SerialException as e:
            raise ConnectionError(f"Error during serial communication: {e}")

    def validate_token(self, token):
        command = f"BUCHI:VALIDATE.TOKEN?{token}\r\n"
        response = self.write_and_read(command)
        return "BUCHI:OK" in response

    def write_url(self, url):
        command = f"BUCHI:WRITE.URL?{url}\r\n"
        response = self.write_and_read(command)
        return response

    def close_port(self):
        if self.port and self.port.is_open:
            self.port.close()
            print("Serial port closed.")

