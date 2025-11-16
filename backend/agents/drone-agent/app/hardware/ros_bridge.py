import logging
import asyncio
from typing import Optional, Callable
import subprocess

logger = logging.getLogger(__name__)


class ROSTopicBridge:
    def __init__(self):
        self.arrival_callback: Optional[Callable] = None
        self.drop_ready_callback: Optional[Callable] = None
        self.home_callback: Optional[Callable] = None
        self._listeners_active = False
        self._tasks = []
        self._last_arrival_time = 0
        self._last_drop_ready_time = 0
        self._last_home_time = 0
    
    def set_arrival_callback(self, callback: Callable):
        self.arrival_callback = callback
        logger.info("Arrival callback registered")
    
    def set_drop_ready_callback(self, callback: Callable):
        self.drop_ready_callback = callback
        logger.info("Drop ready callback registered")
    
    def set_home_callback(self, callback: Callable):
        self.home_callback = callback
        logger.info("Home callback registered")
    
    async def start_listeners(self):
        if self._listeners_active:
            logger.warning("ROS topic listeners already active")
            return
        
        self._listeners_active = True
        
        self._tasks.append(asyncio.create_task(self._listen_arrival()))
        self._tasks.append(asyncio.create_task(self._listen_drop_ready()))
        self._tasks.append(asyncio.create_task(self._listen_home()))
        
        logger.info("ROS topic listeners started successfully")
        logger.info("  - /drone/delivery/arrived")
        logger.info("  - /drone/delivery/drop_ready")
        logger.info("  - /drone/delivery/home_arrived")
    
    async def stop_listeners(self):
        self._listeners_active = False
        for task in self._tasks:
            if not task.done():
                task.cancel()
                try:
                    await task
                except asyncio.CancelledError:
                    pass
        self._tasks.clear()
        logger.info("ROS topic listeners stopped")
    
    async def send_drop_confirmation(self) -> bool:
        try:
            logger.info("Publishing drop confirmation to /drone/delivery/drop_confirm")
            
            result = subprocess.run(
                ['rostopic', 'pub', '-1', '/drone/delivery/drop_confirm', 
                 'std_msgs/Bool', 'data: true'],
                capture_output=True,
                text=True,
                timeout=5
            )
            
            if result.returncode == 0:
                logger.info("Drop confirmation published successfully")
                return True
            else:
                logger.error(f"Failed to publish drop confirmation: {result.stderr}")
                return False
                
        except Exception as e:
            logger.error(f"Error sending drop confirmation: {e}")
            return False
    
    async def _listen_arrival(self):
        logger.info("Started listening to /drone/delivery/arrived")
        
        while self._listeners_active:
            try:
                result = subprocess.run(
                    ['rostopic', 'echo', '-n', '1', '/drone/delivery/arrived'],
                    capture_output=True,
                    text=True,
                    timeout=2
                )
                
                if result.returncode == 0 and result.stdout:
                    current_time = asyncio.get_event_loop().time()
                    if current_time - self._last_arrival_time > 5:
                        self._last_arrival_time = current_time
                        logger.info("Drone arrival detected via ROS topic")
                        if self.arrival_callback:
                            await self.arrival_callback()
                
            except subprocess.TimeoutExpired:
                pass
            except Exception as e:
                logger.error(f"Error listening to arrival topic: {e}")
            
            await asyncio.sleep(0.5)
    
    async def _listen_drop_ready(self):
        logger.info("Started listening to /drone/delivery/drop_ready")
        
        while self._listeners_active:
            try:
                result = subprocess.run(
                    ['rostopic', 'echo', '-n', '1', '/drone/delivery/drop_ready'],
                    capture_output=True,
                    text=True,
                    timeout=2
                )
                
                if result.returncode == 0 and result.stdout:
                    current_time = asyncio.get_event_loop().time()
                    if current_time - self._last_drop_ready_time > 5:
                        self._last_drop_ready_time = current_time
                        logger.info("Cargo drop ready detected via ROS topic")
                        if self.drop_ready_callback:
                            await self.drop_ready_callback()
                
            except subprocess.TimeoutExpired:
                pass
            except Exception as e:
                logger.error(f"Error listening to drop_ready topic: {e}")
            
            await asyncio.sleep(0.5)
    
    async def _listen_home(self):
        logger.info("Started listening to /drone/delivery/home_arrived")
        
        while self._listeners_active:
            try:
                result = subprocess.run(
                    ['rostopic', 'echo', '-n', '1', '/drone/delivery/home_arrived'],
                    capture_output=True,
                    text=True,
                    timeout=2
                )
                
                if result.returncode == 0 and result.stdout:
                    current_time = asyncio.get_event_loop().time()
                    if current_time - self._last_home_time > 5:
                        self._last_home_time = current_time
                        logger.info("Drone home arrival detected via ROS topic")
                        if self.home_callback:
                            await self.home_callback()
                
            except subprocess.TimeoutExpired:
                pass
            except Exception as e:
                logger.error(f"Error listening to home_arrived topic: {e}")
            
            await asyncio.sleep(0.5)
    
    async def get_battery_status(self) -> dict:
        try:
            result = subprocess.run(
                ['rostopic', 'echo', '-n', '1', '/mavros/battery'],
                capture_output=True,
                text=True,
                timeout=5
            )
            
            if result.returncode == 0 and result.stdout:
                lines = result.stdout.strip().split('\n')
                voltage = 12.0
                percentage = 100.0
                
                for line in lines:
                    if 'voltage:' in line:
                        try:
                            voltage = float(line.split(':')[1].strip())
                        except (ValueError, IndexError):
                            pass
                    elif 'percentage:' in line:
                        try:
                            percentage = float(line.split(':')[1].strip()) * 100.0
                        except (ValueError, IndexError):
                            pass
                
                return {
                    'voltage': voltage,
                    'percentage': percentage
                }
        except Exception as e:
            logger.error(f"Error reading battery status: {e}")
        
        return {'voltage': 12.0, 'percentage': 100.0}
    
    async def get_current_pose(self) -> Optional[dict]:
        try:
            result = subprocess.run(
                ['rostopic', 'echo', '-n', '1', '/mavros/local_position/pose'],
                capture_output=True,
                text=True,
                timeout=5
            )
            
            if result.returncode == 0 and result.stdout:
                return {'frame_id': 'map', 'raw_data': result.stdout}
        except Exception as e:
            logger.error(f"Error reading pose: {e}")
        
        return None

ros_bridge = ROSTopicBridge()