from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import logging

from config.config import settings
from app.api import cell_routes, qr_routes
from app.dependencies import cleanup
from app.models.schemas import HealthResponse, ServiceInfoResponse

logging.basicConfig(
    level=settings.log_level,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Starting OrangePI Parcel Automat Service...")
    from app.dependencies import scanner_worker, cell_service

    mapping = cell_service.get_mapping()
    if mapping:
        automat_id = cell_service.cell_repo.get_parcel_automat_id()
        logger.info(
            f"Found existing cell mapping for automat {automat_id} with {len(mapping)} cells")
    else:
        logger.warning(
            "No cell mapping found. Waiting for sync from orchestrator (POST /api/cells/sync)")

    await scanner_worker.start()
    logger.info("QR scanner worker started")
    yield
    logger.info("Shutting down OrangePI Parcel Automat Service...")
    await cleanup()

app = FastAPI(
    title=settings.api_title,
    version=settings.api_version,
    description="OrangePI service for parcel automat cell management and QR scanning",
    lifespan=lifespan
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(cell_routes.router)
app.include_router(qr_routes.router)


@app.get("/", response_model=ServiceInfoResponse)
async def root():
    return ServiceInfoResponse(
        service="OrangePI Parcel Automat Service",
        version=settings.api_version,
        status="running"
    )


@app.get("/health", response_model=HealthResponse)
async def health_check():
    return HealthResponse(status="healthy")

if __name__ == "__main__":
    import uvicorn
    logger.info(f"Starting server on {settings.api_host}:{settings.api_port}")
    uvicorn.run(
        "main:app",
        host=settings.api_host,
        port=settings.api_port,
        reload=False
    )
