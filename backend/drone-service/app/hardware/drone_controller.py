import os
from typing import Dict, Any, Optional
from djitellopy import Tello
import threading
import time

class DroneController:
    def __init__(self):
        self.drone: Optional[Tello] = None
        self.is_connected = False
        self.current_delivery = None
        self.status = "idle"
        self.battery_level = 100.0
        self.position = {"latitude": 0.0, "longitude": 0.0, "altitude": 0.0}
        
    def connect(self) -> bool:
        try:
            drone_ip = os.getenv('DRONE_IP', '192.168.10.1')
            self.drone = Tello(drone_ip)
            self.drone.connect()
            self.is_connected = True
            self.battery_level = self.drone.get_battery()
            return True
        except Exception as e:
            print(f"Failed to connect to drone: {e}")
            return False
    
    def disconnect(self):
        if self.drone and self.is_connected:
            try:
                self.drone.end()
                self.is_connected = False
            except Exception as e:
                print(f"Error disconnecting drone: {e}")
    
    def start_delivery(self, order_id: int, good_id: int, locker_cell_id: int, dimensions: Dict[str, float]) -> Dict[str, Any]:
        if not self.is_connected:
            if not self.connect():
                return {
                    "success": False,
                    "message": "Failed to connect to drone",
                    "delivery_id": 0
                }
        
        self.current_delivery = {
            "order_id": order_id,
            "good_id": good_id,
            "locker_cell_id": locker_cell_id,
            "dimensions": dimensions
        }
        
        self.status = "in_flight"
        
        thread = threading.Thread(target=self._execute_delivery)
        thread.start()
        
        return {
            "success": True,
            "message": "Delivery started",
            "delivery_id": order_id
        }
    
    def _execute_delivery(self):
        try:
            if not self.drone:
                return
            
            self.drone.takeoff()
            time.sleep(2)
            
            self.drone.move_forward(50)
            time.sleep(1)
            
            self.drone.move_down(20)
            time.sleep(1)
            
            self.drone.land()
            
            self.status = "idle"
            self.current_delivery = None
            
        except Exception as e:
            print(f"Error during delivery: {e}")
            self.status = "error"
    
    def get_status(self, drone_id: int) -> Dict[str, Any]:
        battery = self.battery_level
        if self.drone and self.is_connected:
            try:
                battery = self.drone.get_battery()
                self.battery_level = battery
            except:
                pass
        
        return {
            "drone_id": drone_id,
            "status": self.status,
            "battery_level": battery,
            "current_position": self.position
        }
    
    def cancel_delivery(self, delivery_id: int) -> Dict[str, Any]:
        if self.current_delivery and self.current_delivery["order_id"] == delivery_id:
            try:
                if self.drone and self.is_connected:
                    self.drone.emergency()
                self.status = "idle"
                self.current_delivery = None
                return {"success": True, "message": "Delivery cancelled"}
            except Exception as e:
                return {"success": False, "message": f"Failed to cancel: {str(e)}"}
        
        return {"success": False, "message": "Delivery not found"}

