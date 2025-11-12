import asyncio
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI, WebSocket, Request
from fastapi.middleware.cors import CORSMiddleware

from config.config import settings
from app.dependencies import (
    delivery_use_case,
    drone_ws_handler,
    admin_ws_handler,
    rabbitmq_client,
    drone_manager,
    video_proxy_handler,
    init_dependencies,
    cleanup
)
from app.services.workers.delivery import DeliveryWorker
from app.api.prometheus import PrometheusMiddleware, metrics_handler

logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)

delivery_worker = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global delivery_worker
    import logging
    logger = logging.getLogger("drone-service")

    logger.warning("="*60)
    logger.warning("LIFESPAN STARTING - INITIALIZING DEPENDENCIES")
    logger.warning("="*60)

    await init_dependencies()
    logger.warning("Dependencies initialized")

    delivery_worker = DeliveryWorker(rabbitmq_client, delivery_use_case)
    logger.warning("Delivery worker created")

    try:
        logger.warning("Starting delivery worker consume...")
        await delivery_worker.start()
        logger.warning("="*60)
        logger.warning("DELIVERY WORKER STARTED SUCCESSFULLY")
        logger.warning("="*60)
    except Exception as e:
        logger.error(f"FAILED TO START DELIVERY WORKER: {e}")
        import traceback
        traceback.print_exc()

    yield

    logger.warning("Shutting down delivery worker...")
    if delivery_worker:
        await delivery_worker.stop()
    await cleanup()


app = FastAPI(title="Drone Service", version="1.0.0", lifespan=lifespan)

app.add_middleware(PrometheusMiddleware)
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.websocket("/ws/drone")
async def drone_websocket_route(websocket: WebSocket):
    await drone_ws_handler.handle_drone_connection(websocket)


@app.websocket("/ws/admin")
async def admin_websocket_route(websocket: WebSocket):
    await admin_ws_handler.handle_admin_connection(websocket)


@app.websocket("/ws/drone/{drone_id}/video")
async def drone_video_route(websocket: WebSocket, drone_id: str):
    await video_proxy_handler.handle_admin_video_connection(websocket, drone_id)


@app.post("/api/drones/{drone_id}/command")
async def drone_command_route(drone_id: str, command_req: dict):
    from app.api.schemas import DroneCommandRequest

    try:
        req = DroneCommandRequest(**command_req)
    except Exception as e:
        return {"success": False, "message": str(e)}

    if req.command == "return_home":
        success = await drone_ws_handler.send_command_to_drone(
            drone_id,
            {"command": "return_home"}
        )
        if success:
            return {"success": True, "message": "Return home command sent"}
        return {"success": False, "message": "Drone not connected"}

    return {"success": False, "message": "Unknown command"}


@app.get("/metrics")
async def metrics_route(request: Request):
    return await metrics_handler(request)


@app.get("/health")
async def health_check():
    return {"status": "healthy", "service": "drone-service"}


@app.get("/status")
async def get_status():
    try:
        drones = drone_manager.get_all_drones()
        if not drones:
            return {
                "drone_id": "unknown",
                "status": "no_drones_registered",
                "battery_level": 0.0,
                "position": {
                    "latitude": 0.0,
                    "longitude": 0.0,
                    "altitude": 0.0
                },
                "speed": 0.0,
                "current_delivery_id": None,
                "error_message": "No drones registered"
            }

        drone_id = drones[0]
        state = await drone_manager.get_drone_state(drone_id)

        if state:
            return {
                "drone_id": state.drone_id,
                "status": state.status.value,
                "battery_level": state.battery_level,
                "position": {
                    "latitude": state.current_position.latitude,
                    "longitude": state.current_position.longitude,
                    "altitude": state.current_position.altitude
                },
                "speed": state.speed,
                "current_delivery_id": state.current_delivery_id,
                "error_message": state.error_message
            }
        else:
            return {
                "drone_id": drone_id,
                "status": "unknown",
                "battery_level": 0.0,
                "position": {
                    "latitude": 0.0,
                    "longitude": 0.0,
                    "altitude": 0.0
                },
                "speed": 0.0,
                "current_delivery_id": None,
                "error_message": "Drone state not available"
            }
    except Exception as e:
        return {
            "drone_id": "unknown",
            "status": "error",
            "battery_level": 0.0,
            "position": {
                "latitude": 0.0,
                "longitude": 0.0,
                "altitude": 0.0
            },
            "speed": 0.0,
            "current_delivery_id": None,
            "error_message": str(e)
        }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=settings.HTTP_PORT)
