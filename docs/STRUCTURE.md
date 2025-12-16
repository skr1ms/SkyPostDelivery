# Project Structure

## Overview

SkyPost Delivery is a comprehensive drone-based autonomous delivery system consisting of multiple microservices, agents, and client applications. The system is designed for delivering goods to parcel lockers using drones with ArUco marker-based navigation.

## Root Directory Structure

```
SkyPostDelivery/
├── admin_panel/              # React-based administrative web interface
├── backend/                  # Backend services and agents
│   ├── agents/              # Hardware agents (drone, locker)
│   ├── drone-service/       # Drone management and WebSocket service
│   └── go-orchestrator/     # Main orchestration service
├── certificates/            # SSL/TLS certificates
├── deployment/              # Deployment configurations
│   ├── docker/             # Docker Compose files
│   └── nginx/              # Nginx reverse proxy configurations
├── docs/                    # Project documentation
├── mobileapp/              # Flutter mobile application
├── monitoring/             # Monitoring and observability stack
│   ├── grafana/           # Grafana dashboards and provisioning
│   ├── loki/              # Loki log aggregation configuration
│   ├── prometheus/        # Prometheus metrics configuration
│   └── promtail/          # Promtail log collection configuration
├── proto/                  # Protocol Buffers definitions
├── .env.example            # Environment variables template
├── .gitignore              # Git ignore rules
├── LICENSE                 # Project license
├── README.md               # Project overview
└── shell.nix              # Nix development environment
```

## Backend Services

### 1. Go Orchestrator (`backend/go-orchestrator/`)

**Purpose**: Main orchestration service managing orders, users, goods, parcel automats, and deliveries.

**Structure**:
```
go-orchestrator/
├── cmd/
│   └── app/
│       └── main.go                 # Application entry point
├── config/
│   └── config.go                   # Configuration management
├── docs/                           # Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── app/
│   │   └── app.go                  # Application initialization
│   ├── controller/
│   │   ├── grpc/                   # gRPC controllers
│   │   │   ├── orchestrator.go    # Orchestrator gRPC service
│   │   │   └── server.go           # gRPC server setup
│   │   └── http/
│   │       ├── middleware/         # HTTP middlewares
│   │       │   ├── cors.go        # CORS middleware
│   │       │   ├── jwt.go         # JWT authentication
│   │       │   ├── logger.go      # Request logging
│   │       │   └── rate_limiter.go # Rate limiting
│   │       └── v1/                 # API v1 routes
│   │           ├── delivery.go    # Delivery endpoints
│   │           ├── drone.go       # Drone endpoints
│   │           ├── good.go        # Goods endpoints
│   │           ├── locker.go      # Locker cell endpoints
│   │           ├── monitoring.go  # Monitoring endpoints
│   │           ├── order.go       # Order endpoints
│   │           ├── parcel_automat.go # Parcel automat endpoints
│   │           ├── qr.go          # QR code endpoints
│   │           ├── router.go      # Router configuration
│   │           └── user.go        # User authentication endpoints
│   ├── entity/                     # Domain entities
│   │   ├── delivery.go
│   │   ├── drone.go
│   │   ├── good.go
│   │   ├── locker.go
│   │   ├── order.go
│   │   ├── parcel_automat.go
│   │   ├── qr.go
│   │   └── user.go
│   ├── repo/                       # Repository layer
│   │   ├── persistent/            # PostgreSQL repositories
│   │   │   ├── sqlc/             # Generated SQLC code
│   │   │   ├── delivery_postgres.go
│   │   │   ├── drone_postgres.go
│   │   │   ├── good_postgres.go
│   │   │   ├── locker_postgres.go
│   │   │   ├── order_postgres.go
│   │   │   ├── parcel_automat_postgres.go
│   │   │   └── user_postgres.go
│   │   └── webapi/                # External API adapters
│   │       ├── orangepi_adapter.go # Locker agent HTTP client
│   │       ├── qr_adapter.go      # QR code generation
│   │       └── smsaero_api.go     # SMS notification service
│   └── usecase/                    # Business logic layer
│       ├── delivery.go
│       ├── drone.go
│       ├── good.go
│       ├── locker.go
│       ├── notification.go
│       ├── order.go
│       ├── order_worker.go         # Background order processor
│       ├── parcel_automat.go
│       ├── qr.go
│       └── user.go
├── migrations/                     # Database migrations
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── pkg/                            # Shared packages
│   ├── grpc/                      # gRPC client packages
│   ├── jwt/                       # JWT utilities
│   ├── minio/                     # MinIO client
│   ├── postgres/                  # PostgreSQL client
│   ├── qr/                        # QR code generation
│   ├── rabbitmq/                  # RabbitMQ client
│   └── sms/                       # SMS service
├── queries/                        # SQL queries for SQLC
│   ├── deliveries.sql
│   ├── drones.sql
│   ├── goods.sql
│   ├── locker_cells.sql
│   ├── orders.sql
│   ├── parcel_automats.sql
│   └── users.sql
├── schema/
│   └── schema.sql                  # Database schema
├── Dockerfile                      # Docker image definition
├── Makefile                        # Build automation
├── go.mod                          # Go module dependencies
├── go.sum                          # Dependency checksums
└── sqlc.yaml                       # SQLC configuration
```

**Key Responsibilities**:
- User authentication and authorization (JWT)
- Order creation and management
- Parcel automat management
- QR code generation and validation
- SMS notifications (via SMSAero API)
- Delivery orchestration and assignment
- gRPC server for cell opening requests
- HTTP REST API for clients and admin panel
- RabbitMQ message publishing for drone tasks
- Background worker for processing pending orders

### 2. Drone Service (`backend/drone-service/`)

**Purpose**: Manages drone connections, WebSocket communication, video streaming, and delivery execution.

**Structure**:
```
drone-service/
├── cmd/
│   └── app/
│       └── main.go                 # Application entry point
├── config/
│   └── config.go                   # Configuration management
├── docs/                           # Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── app/
│   │   └── app.go                  # Application initialization
│   ├── controller/
│   │   ├── http/
│   │   │   ├── middleware/        # HTTP middlewares
│   │   │   │   ├── cors.go
│   │   │   │   └── logger.go
│   │   │   └── v1/
│   │   │       ├── drone.go       # Drone HTTP endpoints
│   │   │       └── router.go      # Router configuration
│   │   └── websocket/             # WebSocket handlers
│   │       ├── admin_handler.go   # Admin panel WebSocket
│   │       ├── drone_handler.go   # Drone agent WebSocket
│   │       └── video_handler.go   # Video streaming handler
│   ├── entity/                     # Domain entities
│   │   ├── delivery.go
│   │   └── drone.go
│   ├── repo/                       # Repository layer
│   │   ├── contracts.go           # Repository interfaces
│   │   └── persistent/            # PostgreSQL repositories
│   │       ├── sqlc/             # Generated SQLC code
│   │       ├── delivery_postgres.go
│   │       └── drone_postgres.go
│   └── usecase/                    # Business logic layer
│       ├── contracts.go           # Use case interfaces
│       ├── delivery.go            # Delivery execution logic
│       ├── drone_manager.go       # Drone connection management
│       └── drone_message.go       # Drone message processing
├── pkg/                            # Shared packages
│   ├── grpc/                      # gRPC client (to orchestrator)
│   ├── postgres/                  # PostgreSQL client
│   └── rabbitmq/                  # RabbitMQ consumer
├── queries/                        # SQL queries for SQLC
│   ├── deliveries.sql
│   └── drones.sql
├── Dockerfile                      # Docker image definition
├── Makefile                        # Build automation
├── go.mod                          # Go module dependencies
├── go.sum                          # Dependency checksums
└── sqlc.yaml                       # SQLC configuration
```

**Key Responsibilities**:
- WebSocket connection management for drones
- WebSocket connections for admin panel monitoring
- Video frame relay from drones to admin panel
- Drone registration and status tracking
- Delivery task assignment to drones
- RabbitMQ consumer for delivery tasks
- gRPC client to orchestrator for cell opening requests
- Heartbeat and telemetry data collection
- Drone command sending (drop cargo, return to base)

### 3. Agents

#### 3.1 Drone Agent (`backend/agents/drone-agent/`)

**Purpose**: Python-based agent running on Clover drone hardware, manages flight, ROS integration, and WebSocket communication.

**Structure**:
```
drone-agent/
├── app/
│   ├── main.py                     # Application entry point
│   ├── config/
│   │   └── settings.py            # Configuration dataclasses
│   ├── core/                       # Core business logic
│   │   ├── ros_bridge.py          # ROS topic integration
│   │   ├── state_machine.py       # Delivery state management
│   │   └── websocket_client.py    # WebSocket client
│   ├── hardware/                   # Hardware abstractions
│   │   ├── camera_handler.py      # ROS camera integration
│   │   └── flight_controller.py   # Flight script launcher
│   ├── models/                     # Data models
│   │   ├── messages.py            # WebSocket message types
│   │   ├── task.py                # Delivery task model
│   │   └── telemetry.py           # Telemetry data structures
│   ├── services/                   # Application services
│   │   ├── delivery_service.py    # Delivery orchestration
│   │   ├── telemetry_service.py   # Telemetry reporting
│   │   └── video_service.py       # Video streaming
│   └── utils/                      # Utility functions
│       ├── logger.py              # Logging setup
│       └── retry.py               # Retry decorator
├── config/                         # Configuration files
│   └── .env.example               # Environment variables template
├── scripts/                        # Flight scripts
│   ├── delivery_flight.py         # Delivery mission script
│   └── flight_back.py             # Return to base script
├── tests/                          # Unit tests
│   ├── conftest.py                # Test configuration
│   ├── test_delivery_service.py
│   ├── test_state_machine.py
│   ├── test_telemetry_service.py
│   └── test_video_service.py
├── run_as_service/                 # Systemd service files
│   ├── drone-agent.service
│   ├── install.sh
│   ├── uninstall.sh
│   └── README.md
├── .env.example                    # Environment variables template
├── Makefile                        # Development automation
├── pytest.ini                      # Pytest configuration
├── requirements.txt                # Production dependencies
├── requirements-dev.txt            # Development dependencies
└── start_drone.sh                  # Manual start script
```

**Key Responsibilities**:
- WebSocket connection to drone-service
- ROS Noetic integration (topics, services)
- Flight script execution (delivery, return to base)
- ArUco marker navigation
- Cargo drop via servo control (pigpio)
- Camera video streaming (ROS /main_camera/image_raw topic)
- Telemetry reporting (battery, position, status)
- Delivery state machine management
- Event idempotency handling

**Technologies**:
- Python 3.8+
- ROS Noetic
- Clover flight stack
- OpenCV with cv_bridge
- asyncio for concurrent operations
- websockets library

#### 3.2 Locker Agent (`backend/agents/locker-agent/`)

**Purpose**: Go-based agent running on Orange Pi hardware in parcel automats, manages cells, Arduino communication, and QR scanning.

**Structure**:
```
locker-agent/
├── cmd/
│   └── main.go                     # Application entry point
├── config/
│   └── config.yaml                 # Configuration file
├── internal/
│   ├── app/
│   │   └── app.go                  # Application initialization
│   ├── controller/
│   │   └── http/
│   │       ├── middleware/        # HTTP middlewares
│   │       │   └── middleware.go
│   │       └── v1/
│   │           ├── cell.go        # Cell management endpoints
│   │           ├── health.go      # Health check
│   │           ├── qr.go          # QR scanning endpoints
│   │           └── router.go      # Router configuration
│   ├── entity/                     # Domain entities
│   │   ├── cell.go
│   │   └── qr.go
│   ├── hardware/                   # Hardware abstractions
│   │   ├── arduino.go             # Serial communication with Arduino
│   │   ├── contracts.go           # Hardware interfaces
│   │   ├── display.go             # Display controller
│   │   └── qr_camera.go           # QR code camera
│   ├── repo/
│   │   └── inmemory/              # In-memory cell mapping
│   │       ├── cell_mapping.go
│   │       └── contracts.go
│   └── usecase/                    # Business logic layer
│       ├── cell_manager.go        # Cell management use case
│       └── qr_scanner.go          # QR code scanning use case
├── pkg/
│   ├── api/                        # HTTP client for orchestrator
│   │   └── client.go
│   └── logger/                     # Logger interface
│       └── logger.go
├── run_as_service/                 # Systemd service files
│   ├── locker-agent.service
│   ├── install.sh
│   ├── uninstall.sh
│   └── README.md
├── .env.example                    # Environment variables template
├── Dockerfile                      # Docker image definition
├── Makefile                        # Build automation
├── go.mod                          # Go module dependencies
└── go.sum                          # Dependency checksums
```

**Key Responsibilities**:
- Cell synchronization with orchestrator
- Arduino serial communication for cell opening
- QR code validation via orchestrator API
- Display management for user feedback
- Cell mapping (external/internal cell IDs)
- HTTP API for cell opening requests
- Health checks and status reporting

**Hardware Integration**:
- Arduino (serial /dev/ttyUSB0) for cell lock control
- Camera for QR code scanning
- Display for user interface
- Orange Pi as main controller

## Frontend Applications

### 1. Admin Panel (`admin_panel/`)

**Purpose**: React-based web interface for system administrators.

**Structure**:
```
admin_panel/
├── src/
│   ├── App.tsx                     # Main application component
│   ├── main.tsx                    # Application entry point
│   ├── index.css                   # Global styles
│   ├── vite-env.d.ts              # Vite type definitions
│   ├── api/
│   │   └── index.ts               # API client configuration
│   ├── assets/
│   │   └── icon/                  # Application icons
│   ├── components/                 # Reusable components
│   │   ├── ProtectedRoute.tsx     # Route protection
│   │   ├── Button/
│   │   ├── ConfirmModal/
│   │   ├── Layout/
│   │   └── Modal/
│   ├── config/
│   │   └── api_config.ts          # API configuration
│   ├── context/
│   │   └── AuthContext.tsx        # Authentication context
│   ├── pages/                      # Application pages
│   │   ├── Dashboard/             # Dashboard page
│   │   ├── Drones/                # Drone management
│   │   ├── Goods/                 # Goods management
│   │   ├── Login/                 # Login page
│   │   ├── Monitoring/            # Real-time monitoring
│   │   └── ParcelAutomats/        # Parcel automat management
│   ├── types/
│   │   └── index.ts               # TypeScript type definitions
│   └── utils/
│       ├── formatters.ts          # Data formatters
│       └── toast.ts               # Toast notifications
├── public/                         # Static assets
├── index.html                      # HTML template
├── package.json                    # NPM dependencies
├── tsconfig.json                   # TypeScript configuration
├── vite.config.ts                  # Vite configuration
└── Dockerfile                      # Docker image definition
```

**Key Features**:
- Drone monitoring and management
- Goods catalog management
- Parcel automat configuration
- Order tracking
- Real-time delivery monitoring
- Video streaming from drones
- System status dashboard

**Technologies**:
- React 18
- TypeScript
- Vite
- React Router
- Context API for state management
- WebSocket for real-time updates

### 2. Mobile App (`mobileapp/`)

**Purpose**: Flutter-based mobile application for end users.

**Structure**:
```
mobileapp/
├── lib/
│   ├── main.dart                   # Application entry point
│   ├── core/                       # Core utilities
│   │   ├── di/                    # Dependency injection
│   │   ├── services/              # Core services
│   │   │   ├── connectivity_service.dart
│   │   │   └── push_notification_service.dart
│   │   └── theme/                 # App theming
│   │       └── app_theme.dart
│   └── features/                   # Feature modules
│       ├── auth/                  # Authentication feature
│       │   ├── data/
│       │   │   ├── datasources/
│       │   │   ├── models/
│       │   │   └── repositories/
│       │   ├── domain/
│       │   │   ├── entities/
│       │   │   ├── repositories/
│       │   │   └── usecases/
│       │   └── presentation/
│       │       ├── providers/
│       │       ├── screens/
│       │       └── widgets/
│       ├── goods/                 # Goods catalog feature
│       ├── home/                  # Home screen feature
│       ├── orders/                # Order management feature
│       ├── parcel_automats/       # Parcel automat selection
│       ├── profile/               # User profile feature
│       └── qr/                    # QR code feature
├── assets/                         # Static assets
│   └── icon/                      # Application icons
├── android/                        # Android platform files
├── ios/                            # iOS platform files
├── test/                           # Unit tests
├── pubspec.yaml                    # Flutter dependencies
├── analysis_options.yaml           # Dart analyzer configuration
└── flake.nix                       # Nix development environment
```

**Key Features**:
- User authentication (phone/email)
- Goods catalog browsing
- Order creation
- Parcel automat selection
- QR code generation for pickup
- Order tracking
- Push notifications (Firebase)
- Profile management

**Technologies**:
- Flutter 3.x
- Dart
- Provider for state management
- Firebase Cloud Messaging
- Clean Architecture pattern

## Deployment Configuration

### Docker Compose Files (`deployment/docker/`)

```
docker/
├── docker-compose.dev.yml          # Development environment
├── docker-compose.prod.yml         # Production environment
├── docker-compose.cicd.yml         # CI/CD pipeline
└── docker-compose.npm.yml          # NPM dependency cache
```

**Services Defined**:
- PostgreSQL database
- MinIO object storage
- Redis cache
- RabbitMQ message broker
- go-orchestrator service
- drone-service
- admin-panel
- Nginx reverse proxy
- Prometheus
- Grafana
- Loki
- Promtail
- cAdvisor
- Node Exporter
- Postgres Exporter

### Nginx Configuration (`deployment/nginx/`)

```
nginx/
├── dev.conf                        # Development proxy configuration
└── prod.conf                       # Production proxy configuration
```

**Routes**:
- `/api/v1/` → go-orchestrator HTTP API
- `/ws/` → drone-service WebSocket
- `/admin/` → admin-panel static files
- `/swagger/` → API documentation

## Monitoring Stack (`monitoring/`)

### Prometheus Configuration
```
prometheus/
├── prometheus.yml                  # Main configuration
└── alerts.yml                      # Alert rules
```

**Scrape Targets**:
- go-orchestrator (port 9091)
- drone-service (port 9092)
- RabbitMQ (port 15692)
- PostgreSQL (via postgres-exporter)
- cAdvisor (container metrics)
- Node Exporter (host metrics)

### Grafana Configuration
```
grafana/
├── dashboards/                     # Dashboard JSON files
│   ├── system-overview.json
│   ├── service-metrics.json
│   └── drone-monitoring.json
└── provisioning/                   # Provisioning configuration
    ├── dashboards/
    │   └── dashboards.yml
    └── datasources/
        └── datasources.yml
```

**Dashboards**:
- System Overview: Overall system health
- Service Metrics: Service-specific metrics
- Drone Monitoring: Real-time drone tracking
- RabbitMQ Dashboard: Message queue metrics

### Loki Configuration
```
loki/
└── loki-config.yml                 # Log aggregation configuration
```

### Promtail Configuration
```
promtail/
└── promtail-config.yml            # Log collection configuration
```

## Protocol Buffers (`proto/`)

```
proto/
└── orchestrator.proto              # gRPC service definitions
```

**Services Defined**:
- `OrchestratorService`: Cell opening requests

## Development Tools

### Nix Environment (`shell.nix`)

Provides consistent development environment with:
- Go 1.21
- Python 3.13
- Node.js 20
- Flutter SDK
- Docker and docker-compose
- PostgreSQL client tools
- Protocol Buffer compiler

### Makefiles

Each service includes a Makefile for:
- Building binaries
- Running tests
- Code generation (SQLC, mocks)
- Docker image building
- Linting and formatting

## Configuration Files

### Environment Variables

Each component has `.env.example` files documenting required environment variables:

**Orchestrator Variables**:
- Database connection
- JWT secrets
- MinIO configuration
- SMS service credentials
- RabbitMQ connection
- gRPC endpoints

**Drone Service Variables**:
- Database connection
- WebSocket configuration
- RabbitMQ connection
- gRPC client configuration

**Agent Variables**:
- WebSocket endpoint
- Hardware configuration
- Log levels
- Service URLs

## CI/CD Configuration

### GitLab CI Files

```
.gitlab-ci.yml                      # Main CI/CD pipeline
.gitlab-ci-dev.yml                  # Development pipeline
.gitlab-ci-prod.yml                 # Production pipeline
```

**Pipeline Stages**:
1. Build: Compile all services
2. Test: Run unit and integration tests
3. Deploy: Deploy to target environment

## Documentation Structure

```
docs/
├── ARCHITECTURE.md                 # System architecture
├── API-SPEC.md                     # API specifications
├── DATABASE.md                     # Database documentation
├── DEPLOYMENT.md                   # Deployment guide
├── MONITORING.md                   # Monitoring guide
└── STRUCTURE.md                    # Project structure (this file)
```

## Summary

This project structure follows these principles:

1. **Microservices Architecture**: Independent services with clear boundaries
2. **Clean Architecture**: Separation of concerns within each service
3. **Hexagonal Architecture**: Ports and adapters pattern for agents
4. **Domain-Driven Design**: Clear domain models and use cases
5. **Repository Pattern**: Data access abstraction
6. **Dependency Injection**: Loose coupling between components
7. **Configuration Management**: Environment-based configuration
8. **Observability**: Comprehensive monitoring and logging
9. **Containerization**: Docker-based deployment
10. **Code Generation**: SQLC for type-safe SQL queries

The structure supports:
- Independent development and deployment of services
- Clear separation between business logic and infrastructure
- Easy testing and mocking
- Scalability and maintainability
- Multiple client platforms (web, mobile)
- Hardware agent integration
- Real-time monitoring and debugging
