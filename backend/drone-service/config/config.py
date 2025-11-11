import os
from dotenv import load_dotenv

load_dotenv()


class Settings:
    GO_GRPC_PORT: str = os.getenv("PYTHON_GRPC_PORT", "50051")
    HTTP_PORT: int = int(os.getenv("DRONE_SERVICE_HTTP_PORT", "8081"))
    
    DATABASE_URL: str = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/hitech")
    
    DRONE_SERVICE_HTTP_URL: str = os.getenv("DRONE_SERVICE_HTTP_URL", "http://192.168.1.100:8080")
    
    ORCHESTRATOR_GRPC_URL: str = os.getenv("ORCHESTRATOR_GRPC_URL", "localhost:50052")
    RABBITMQ_URL: str = os.getenv("RABBITMQ_URL", "amqp://admin:admin@localhost:5672/")
    
    MINIO_ENDPOINT: str = os.getenv("MINIO_ENDPOINT", "localhost:9000")
    MINIO_ROOT_USER: str = os.getenv("MINIO_ROOT_USER", "admin")
    MINIO_ROOT_PASSWORD: str = os.getenv("MINIO_ROOT_PASSWORD", "admin")
    MINIO_USE_SSL: bool = os.getenv("MINIO_USE_SSL", "false").lower() == "true"
    MINIO_BUCKET_RECORDS: str = os.getenv("MINIO_BUCKET_RECORDS", "records")
    
    WEBSOCKET_BROADCAST_INTERVAL: int = int(os.getenv("DRONE_SERVICE_WEBSOCKET_INTERVAL", "5"))
    
    LOG_LEVEL: str = os.getenv("LOG_LEVEL", "INFO")


settings = Settings()

