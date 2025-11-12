import subprocess
import logging
from typing import Optional
from pathlib import Path

logger = logging.getLogger(__name__)


class FlightScriptManager:
    def __init__(self, scripts_dir: str = "/root"):
        self.scripts_dir = Path(scripts_dir)
        self.current_process: Optional[subprocess.Popen] = None
        self.delivery_script = self.scripts_dir / "delivery_flight.py"
        self.return_script = self.scripts_dir / "flight_back.py"
    
    async def launch_delivery_flight(self, aruco_id: int, home_aruco_id: int = 131) -> bool:
        try:
            if self.current_process and self.current_process.poll() is None:
                logger.warning("Flight script already running")
                return False
            
            cmd = [
                "python3",
                str(self.delivery_script),
                str(aruco_id),
                str(home_aruco_id)
            ]
            
            logger.info(f"Launching delivery flight: {' '.join(cmd)}")
            
            self.current_process = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                cwd=str(self.scripts_dir)
            )
            
            logger.info(f"Delivery flight script launched with PID {self.current_process.pid}")
            return True
            
        except Exception as e:
            logger.error(f"Failed to launch delivery flight: {e}")
            return False
    
    async def launch_return_flight(self, home_aruco_id: int = 131, home_x: float = 0.0, home_y: float = 0.0) -> bool:
        try:
            cmd = [
                "python3",
                str(self.return_script),
                str(home_aruco_id),
                str(home_x),
                str(home_y)
            ]
            
            logger.info(f"Launching return flight: {' '.join(cmd)}")
            
            return_process = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                cwd=str(self.scripts_dir)
            )
            
            logger.info(f"Return flight script launched with PID {return_process.pid}")
            return True
            
        except Exception as e:
            logger.error(f"Failed to launch return flight: {e}")
            return False
    
    def is_flight_active(self) -> bool:
        if self.current_process is None:
            return False
        return self.current_process.poll() is None
    
    def terminate_current_flight(self):
        if self.current_process and self.current_process.poll() is None:
            logger.info("Terminating current flight script")
            self.current_process.terminate()
            try:
                self.current_process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                logger.warning("Flight script did not terminate, killing")
                self.current_process.kill()
            self.current_process = None


flight_manager = FlightScriptManager()
