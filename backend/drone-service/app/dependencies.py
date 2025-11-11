from app.models.ports import StateRepositoryPort
from app.repositories.postgres import PostgresStateRepository
from app.services.drone_manager import DroneManager
from app.services.use_cases import DeliveryUseCase
from app.api.handlers.drone_websocket import DroneWebSocketHandler
from app.api.handlers.admin_websocket import AdminWebSocketHandler
from app.services.video_service import DroneVideoProxyHandler
from app.hardware.rabbitmq import RabbitMQClient
from app.hardware.grpc_client import OrchestratorGRPCClient
from config.config import settings

state_repository: StateRepositoryPort = PostgresStateRepository()

drone_manager = DroneManager(state_repository)
orchestrator_grpc_client = OrchestratorGRPCClient(settings.ORCHESTRATOR_GRPC_URL)
rabbitmq_client = RabbitMQClient(settings.RABBITMQ_URL)

drone_ws_handler = DroneWebSocketHandler(state_repository, drone_manager)
video_proxy_handler = DroneVideoProxyHandler(drone_ws_handler)
drone_ws_handler.video_proxy_handler = video_proxy_handler

admin_ws_handler = AdminWebSocketHandler(state_repository, drone_manager)

delivery_use_case = DeliveryUseCase(
    state_repository=state_repository,
    drone_manager=drone_manager,
    drone_ws_handler=drone_ws_handler,
    orchestrator_grpc_client=orchestrator_grpc_client,
    rabbitmq_client=rabbitmq_client
)

drone_ws_handler.delivery_use_case = delivery_use_case

async def init_dependencies():
    await state_repository.connect()
    if hasattr(rabbitmq_client, 'connect'):
        await rabbitmq_client.connect()

async def cleanup():
    await state_repository.close()
    if hasattr(rabbitmq_client, 'close'):
        await rabbitmq_client.close()
