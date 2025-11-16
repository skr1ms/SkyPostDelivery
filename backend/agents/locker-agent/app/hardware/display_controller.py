import logging
import serial
import time
from typing import Optional
from config.config import settings

logger = logging.getLogger(__name__)


class DisplayController:
    def __init__(self, port: str = None, baudrate: int = 115200):
        self.port = port or getattr(settings, 'display_port', '/dev/ttyUSB0')
        self.baudrate = baudrate
        self.serial_connection: Optional[serial.Serial] = None
        self._mock_mode = False

        if not self.port:
            logger.warning("Display port not configured, running in MOCK mode")
            self._mock_mode = True
        else:
            self._initialize_serial()

    def _initialize_serial(self):
        try:
            import os
            if not os.path.exists(self.port):
                logger.error(f"Display port {self.port} does not exist")
                self._mock_mode = True
                return

            self.serial_connection = serial.Serial(
                port=self.port,
                baudrate=self.baudrate,
                timeout=2.0,
                dsrdtr=False,  
                rtscts=False
            )
            time.sleep(0.5)
            logger.info(
                f"Display connected on {self.port} at {self.baudrate} baud")
        except Exception as e:
            logger.error(f"Failed to connect to display: {e}")
            logger.warning("Running in MOCK mode")
            self._mock_mode = True

    def send_message(self, message: str) -> bool:
        if self._mock_mode:
            logger.info(f"[DISPLAY MOCK] {message}")
            return True

        if not self.serial_connection or not self.serial_connection.is_open:
            logger.error("Display serial connection is not open")
            return False

        try:
            formatted = message + "\n"
            self.serial_connection.write(formatted.encode('utf-8'))
            self.serial_connection.flush()
            logger.info(f"[DISPLAY] Sent: {message}")
            return True
        except Exception as e:
            logger.error(f"Failed to send message to display: {e}")
            return False

    def show_welcome(self):
        self.send_message("Welcome!\nScan your QR")

    def show_scanning(self):
        self.send_message("Scanning...\nPlease wait")

    def show_qr_success(self, customer_name: str = None):
        if customer_name:
            self.send_message(f"Hello,\n{customer_name}!")
        else:
            self.send_message("QR accepted!\nOpening cell...")

    def show_qr_invalid(self):
        self.send_message("Invalid QR!\nTry again")

    def show_cell_opening(self, cell_number: int, order_number: str = None):
        if order_number:
            self.send_message(f"Order #{order_number}\nCell #{cell_number}")
        else:
            self.send_message(f"Your cell:\n#{cell_number}")

    def show_cell_opened(self, cell_number: int):
        self.send_message(f"Cell #{cell_number}\nOPEN")

    def show_please_close(self):
        self.send_message("Take your item\nClose the door")

    def show_cell_closed(self):
        self.send_message("Thank you!\nHave a nice day")

    def show_delivery_incoming(self, order_number: str = None):
        if order_number:
            self.send_message(f"Delivery\nOrder #{order_number}")
        else:
            self.send_message("Delivery\nIncoming...")

    def show_drone_landing(self):
        self.send_message("Drone\nLanding...")

    def show_loading_cell(self, cell_number: int):
        self.send_message(f"Loading\nCell #{cell_number}")

    def show_delivery_complete(self, cell_number: int):
        self.send_message(f"Delivered!\nCell #{cell_number}")

    def show_error(self, message: str = "Error occurred"):
        self.send_message(f"ERROR:\n{message}")

    def show_maintenance(self):
        self.send_message("Maintenance\nmode")

    def clear(self):
        self.send_message("                \n                ")

    def close(self):
        if self.serial_connection and not self._mock_mode:
            try:
                self.serial_connection.close()
                logger.info("Display connection closed")
            except:
                pass
