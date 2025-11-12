import logging
from .services.websocket_service import WebSocketService
from .services.delivery_service import DeliveryService
from .models.schemas import DeliveryTaskPayload
from .hardware.flight_manager import flight_manager
from .hardware.ros_bridge import ros_bridge

logger = logging.getLogger(__name__)

delivery_service = DeliveryService()


async def handle_delivery_task(payload: dict):
    try:
        task = DeliveryTaskPayload(**payload)
        logger.info(f"Delivery task {task.delivery_id}")
        logger.info(f"Good: {task.good_id}, Target: {task.parcel_automat_id}, Marker: {task.aruco_id}")
        
        logger.info(f"Launching flight script to ArUco marker {task.aruco_id}")
        
        success = await flight_manager.launch_delivery_flight(
            aruco_id=task.aruco_id,
            home_aruco_id=131
        )
        
        if success:
            logger.info("Flight script launched successfully")
        else:
            logger.error("Failed to launch flight script")
            
    except Exception as e:
        logger.error(f"Error processing delivery: {e}")


websocket_service = WebSocketService(on_delivery_task=handle_delivery_task)
websocket_service.ros_bridge = ros_bridge

delivery_service.set_status_callback(websocket_service.send_status_update)
delivery_service.set_delivery_update_callback(websocket_service.send_delivery_update)

websocket_service.delivery_service = delivery_service

async def cleanup():
    await websocket_service.close()
    await delivery_service.shutdown()