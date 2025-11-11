import logging
from typing import Optional
import serial

logger = logging.getLogger(__name__)


class ArduinoController:
    def __init__(self, port: str = None, baudrate: int = 9600, timeout: float = 1.0, mock_mode: bool = False):
        self.port = port
        self.baudrate = baudrate
        self.timeout = timeout
        self.serial_connection = None
        self._mock_mode = mock_mode or (port is None)
        self._mock_cells_state = {}

        if self._mock_mode:
            logger.info("Running Arduino controller in MOCK mode")
        else:
            self._initialize_serial()

    def _initialize_serial(self):
        try:
            self.serial_connection = serial.Serial(
                port=self.port,
                baudrate=self.baudrate,
                timeout=self.timeout
            )
            logger.info(f"Connected to Arduino on {self.port}")
        except Exception as e:
            logger.error(f"Failed to connect to Arduino: {e}")
            logger.warning("Running in MOCK mode")
            self._mock_mode = True

    def _send_command(self, command: str) -> str:
        if self._mock_mode:
            logger.info(f"MOCK: Sending to Arduino: {command}")
            if command.startswith("cells_"):
                return "4"
            elif command.startswith("open_"):
                cell_num = command.split("_")[1]
                self._mock_cells_state[cell_num] = "opened"
                return "OK"
            elif command.startswith("close_"):
                cell_num = command.split("_")[1]
                self._mock_cells_state[cell_num] = "closed"
                return "OK"
            elif command.startswith("status_"):
                cell_num = command.split("_")[1]
                return self._mock_cells_state.get(cell_num, "closed")
            return "OK"

        try:
            cmd = f"{command}\n"
            self.serial_connection.write(cmd.encode())
            logger.info(f"Sent to Arduino: {command}")
            response = self.serial_connection.readline().decode().strip()
            logger.info(f"Arduino response: {response}")
            return response
        except Exception as e:
            logger.error(f"Failed to communicate with Arduino: {e}")
            return f"ERROR: {str(e)}"

    def open_cell(self, cell_number: int) -> str:
        command = f"open_{cell_number}"
        response = self._send_command(command)
        logger.info(f"Opened cell {cell_number}, response: {response}")
        return response

    def close_cell(self, cell_number: int) -> str:
        command = f"close_{cell_number}"
        response = self._send_command(command)
        logger.info(f"Closed cell {cell_number}, response: {response}")
        return response

    def get_cell_status(self, cell_number: int) -> str:
        command = f"status_{cell_number}"
        response = self._send_command(command)
        logger.info(f"Status of cell {cell_number}: {response}")
        return response

    def get_cells_count(self) -> Optional[int]:
        command = "cells_0"
        response = self._send_command(command)

        try:
            count = int(response)
            logger.info(f"Arduino cells count: {count}")
            return count
        except ValueError:
            logger.warning(
                f"Invalid cells count response from Arduino: {response}")
            return None

    def close(self):
        if self.serial_connection and not self._mock_mode:
            self.serial_connection.close()
            logger.info("Serial connection closed")
