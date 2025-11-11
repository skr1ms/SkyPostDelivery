.PHONY: help install proto run dev docker-build docker-run clean test lint format

help:
	@echo "Доступные команды:"
	@echo "  make install       - Установить зависимости"
	@echo "  make proto         - Сгенерировать protobuf файлы"
	@echo "  make run           - Запустить сервис"
	@echo "  make dev           - Запустить в dev режиме"
	@echo "  make docker-build  - Собрать Docker образ"
	@echo "  make docker-run    - Запустить в Docker"
	@echo "  make test          - Запустить тесты"
	@echo "  make lint          - Проверить код"
	@echo "  make format        - Отформатировать код"
	@echo "  make clean         - Очистить временные файлы"

install:
	pip install -r requirements.txt

proto:
	mkdir -p proto
	pip install grpcio-tools
	python -m grpc_tools.protoc \
		-I../../proto \
		--python_out=./proto \
		--grpc_python_out=./proto \
		--pyi_out=./proto \
		../../proto/orchestrator.proto
	sed -i 's/^import \(.*\)_pb2 as/from . import \1_pb2 as/' proto/*_pb2_grpc.py
	touch proto/__init__.py

run: proto
	python main.py

dev: proto
	uvicorn main:app --host 0.0.0.0 --port 8000 --reload

docker-build:
	docker build -f Dockerfile -t drone-service:latest .

docker-run:
	docker run -d \
		-p 50051:50051 \
		-p 8000:8000 \
		--name drone-service \
		--env-file .env \
		drone-service:latest

docker-stop:
	docker stop drone-service || true
	docker rm drone-service || true

test: proto
	pytest tests/ -v --asyncio-mode=auto

lint:
	flake8 app/ --max-line-length=100
	mypy app/

format:
	black app/ config/ main.py
	isort app/ config/ main.py

clean:
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	find . -type f -name "*.pyo" -delete
	find . -type d -name "*.egg-info" -exec rm -rf {} + 2>/dev/null || true
	rm -rf .pytest_cache .mypy_cache .coverage htmlcov
	@echo "✅ Временные файлы удалены"

