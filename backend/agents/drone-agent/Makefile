.PHONY: help install install-dev test test-verbose test-coverage lint format clean venv

PYTHON := python3
VENV := venv
VENV_BIN := $(VENV)/bin
PIP := $(VENV_BIN)/pip
PYTEST := $(VENV_BIN)/pytest
RUFF := $(VENV_BIN)/ruff

help:
	@echo "Available commands:"
	@echo "  make venv           - Create virtual environment"
	@echo "  make install        - Install production dependencies"
	@echo "  make install-dev    - Install all dependencies including dev tools"
	@echo "  make test           - Run tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linter (ruff)"
	@echo "  make format         - Format code with ruff"
	@echo "  make clean          - Remove virtual environment and cache files"

venv:
	@echo "Creating virtual environment..."
	$(PYTHON) -m venv $(VENV)
	@echo "Virtual environment created at $(VENV)"
	@echo "To activate: source $(VENV_BIN)/activate"

install: venv
	@echo "Installing production dependencies..."
	$(PIP) install --upgrade pip
	$(PIP) install -r requirements.txt
	@echo "Production dependencies installed"

install-dev: install
	@echo "Installing development dependencies..."
	$(PIP) install -r requirements-dev.txt
	@echo "All dependencies installed"

test:
	@if [ ! -d "$(VENV)" ]; then \
		echo "Virtual environment not found. Run 'make install-dev' first"; \
		exit 1; \
	fi
	@echo "Running tests..."
	$(PYTEST) tests/ -v

test-verbose:
	@if [ ! -d "$(VENV)" ]; then \
		echo "Virtual environment not found. Run 'make install-dev' first"; \
		exit 1; \
	fi
	@echo "Running tests with verbose output..."
	$(PYTEST) tests/ -vv -s

test-coverage:
	@if [ ! -d "$(VENV)" ]; then \
		echo "Virtual environment not found. Run 'make install-dev' first"; \
		exit 1; \
	fi
	@echo "Running tests with coverage..."
	$(PYTEST) tests/ --cov=app --cov-report=term-missing --cov-report=html
	@echo "Coverage report generated in htmlcov/index.html"

lint:
	@if [ ! -d "$(VENV)" ]; then \
		echo "Virtual environment not found. Run 'make install-dev' first"; \
		exit 1; \
	fi
	@echo "Running linter..."
	$(RUFF) check app/ tests/

format:
	@if [ ! -d "$(VENV)" ]; then \
		echo "Virtual environment not found. Run 'make install-dev' first"; \
		exit 1; \
	fi
	@echo "Formatting code..."
	$(RUFF) check --fix app/ tests/
	$(RUFF) format app/ tests/

clean:
	@echo "Cleaning up..."
	rm -rf $(VENV)
	rm -rf .pytest_cache
	rm -rf .coverage
	rm -rf htmlcov
	rm -rf .ruff_cache
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
	@echo "Cleanup complete"
