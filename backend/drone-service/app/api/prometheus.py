import time
from prometheus_client import Counter, Histogram, generate_latest, CONTENT_TYPE_LATEST
from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.responses import Response

http_requests_total = Counter(
    'http_requests_total',
    'Total number of HTTP requests',
    ['method', 'endpoint', 'status']
)

http_request_duration_seconds = Histogram(
    'http_request_duration_seconds',
    'HTTP request duration in seconds',
    ['method', 'endpoint']
)

grpc_requests_total = Counter(
    'grpc_requests_total',
    'Total number of gRPC requests',
    ['method', 'status']
)

grpc_request_duration_seconds = Histogram(
    'grpc_request_duration_seconds',
    'gRPC request duration in seconds',
    ['method']
)

class PrometheusMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        start_time = time.time()

        if request.url.path == '/metrics':
            response = await call_next(request)
            return response

        response = await call_next(request)
        
        duration = time.time() - start_time
        method = request.method
        endpoint = request.url.path
        status = response.status_code

        http_requests_total.labels(method=method, endpoint=endpoint, status=str(status)).inc()
        http_request_duration_seconds.labels(method=method, endpoint=endpoint).observe(duration)

        return response

async def metrics_handler(request: Request):
    return Response(content=generate_latest(), media_type=CONTENT_TYPE_LATEST)
