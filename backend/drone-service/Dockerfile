FROM python:3.14-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY backend/drone-service/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY backend/drone-service/ .
COPY proto/ /tmp/proto/

RUN python -m grpc_tools.protoc \
    -I/tmp/proto \
    --python_out=./proto \
    --grpc_python_out=./proto \
    --pyi_out=./proto \
    /tmp/proto/orchestrator.proto && \
    sed -i 's/^import orchestrator_pb2/from . import orchestrator_pb2/g' proto/orchestrator_pb2_grpc.py

EXPOSE 50051 8000

CMD ["python", "main.py"]

