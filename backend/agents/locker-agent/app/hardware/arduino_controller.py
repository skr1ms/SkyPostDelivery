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
        self._mock_internal_state = {}

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
            if command == "cells":
                return "3"
            elif command.startswith("internal_"):
                door_num = command.split("_")[1]
                self._mock_internal_state[door_num] = "opened"
                return "OK"
            elif command.startswith("open_"):
                cell_num = command.split("_")[1]
                self._mock_cells_state[cell_num] = "opened"
                return "OK"
            elif command == "reset":
                self._mock_cells_state.clear()
                self._mock_internal_state.clear()
                return "OK"
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
        logger.info(f"Close cell {cell_number} requested - Arduino handles this automatically")
        return "OK"

    def get_cell_status(self, cell_number: int) -> str:
        logger.info(f"Status of cell {cell_number} - not supported by Arduino, returning 'closed'")
        return "closed"

    def get_cells_count(self) -> Optional[int]:
        command = "cells"
        response = self._send_command(command)

        try:
            count = int(response)
            logger.info(f"Arduino cells count: {count}")
            return count
        except ValueError:
            logger.warning(
                f"Invalid cells count response from Arduino: {response}")
            return None

    def open_internal_door(self, door_number: int) -> str:
        command = f"internal_{door_number}"
        response = self._send_command(command)
        logger.info(f"Opened internal door {door_number}, response: {response}")
        return response

    def close_internal_door(self, door_number: int) -> str:
        logger.info(f"Close internal door {door_number} requested - Arduino handles this automatically")
        return "OK"

    def get_internal_status(self, door_number: int) -> str:
        logger.info(f"Status of internal door {door_number} - not supported by Arduino, returning 'closed'")
        return "closed"

    def get_internal_cells_count(self) -> Optional[int]:
        logger.info("Internal cells count - returning hardcoded 3 (same as external)")
        return 3

    def close(self):
        if self.serial_connection and not self._mock_mode:
            self.serial_connection.close()
            logger.info("Serial connection closed")
