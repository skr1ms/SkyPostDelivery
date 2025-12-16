# SkyPost Delivery - Monitoring Documentation

## Table of Contents

1. [Overview](#overview)
2. [Monitoring Stack Architecture](#monitoring-stack-architecture)
3. [Prometheus](#prometheus)
   - [Configuration](#prometheus-configuration)
   - [Metrics Collection](#metrics-collection)
   - [Custom Metrics](#custom-metrics)
4. [Grafana](#grafana)
   - [Dashboards](#grafana-dashboards)
   - [Data Sources](#data-sources)
   - [Visualization](#visualization)
5. [Loki & Promtail](#loki--promtail)
   - [Log Aggregation](#log-aggregation)
   - [Log Queries](#log-queries)
6. [Alerting](#alerting)
   - [Alert Rules](#alert-rules)
   - [Alert Channels](#alert-channels)
7. [Exporters](#exporters)
8. [Container Monitoring](#container-monitoring)
9. [Best Practices](#best-practices)
10. [Troubleshooting](#troubleshooting)

---

## Overview

SkyPost Delivery uses a comprehensive monitoring stack based on Prometheus, Grafana, and Loki to provide real-time observability into system performance, application health, and operational metrics.

**Monitoring Components**:
- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization and dashboards
- **Loki**: Log aggregation
- **Promtail**: Log collector
- **cAdvisor**: Container metrics
- **Node Exporter**: Host machine metrics
- **PostgreSQL Exporter**: Database metrics
- **RabbitMQ**: Built-in Prometheus exporter

**Key Features**:
- Real-time metrics collection (15s intervals)
- Automated alerting for critical issues
- Centralized log aggregation
- Pre-configured dashboards for all services
- Container and host-level monitoring
- Database performance tracking
- Message queue monitoring

**Access URLs** (Development):
- Grafana: `http://localhost:3000` (admin/admin)
- Prometheus: `http://localhost:9090`
- Loki: `http://localhost:3100`

---

## Monitoring Stack Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Monitoring Stack                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐         ┌──────────────┐                      │
│  │   Grafana    │◄────────│  Prometheus  │                      │
│  │  (Port 3000) │         │  (Port 9090) │                      │
│  └──────┬───────┘         └──────▲───────┘                      │
│         │                        │                              │
│         │                        │ Scrape Metrics (15s)         │
│         │                        │                              │
│  ┌──────▼───────┐         ┌──────┴─────────────────┐            │
│  │     Loki     │◄────────│  Service Exporters:    │            │
│  │  (Port 3100) │         │  - Go-Orchestrator     │            │
│  └──────▲───────┘         │  - Drone-Service       │            │
│         │                 │  - RabbitMQ            │            │
│         │                 │  - PostgreSQL Exporter │            │
│         │                 │  - Node Exporter       │            │
│         │                 │  - cAdvisor            │            │
│  ┌──────┴───────┐         └────────────────────────┘            │
│  │   Promtail   │                                               │
│  │  (Port 9080) │                                               │
│  └──────▲───────┘                                               │
│         │                                                       │
│         │ Collect Logs                                          │
│         │                                                       │
│  ┌──────┴──────────────────────────────────────────┐            │
│  │         Docker Container Logs                   │            │
│  │  /var/lib/docker/containers/*/*.log             │            │
│  └─────────────────────────────────────────────────┘            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Data Flow**:
1. **Metrics Collection**: Prometheus scrapes metrics from all service endpoints every 15 seconds
2. **Log Collection**: Promtail reads Docker container logs and forwards to Loki
3. **Visualization**: Grafana queries Prometheus and Loki for dashboards
4. **Alerting**: Prometheus evaluates alert rules and triggers notifications

---

## Prometheus

### Prometheus Configuration

**Configuration File**: `monitoring/prometheus/prometheus.yml`

```yaml
global:
  scrape_interval: 15s      # Scrape targets every 15 seconds
  evaluation_interval: 15s  # Evaluate rules every 15 seconds

alerting:
  alertmanagers:
    - static_configs:
        - targets: []

rule_files:
  - /etc/prometheus/alerts.yml

scrape_configs:
  # Prometheus self-monitoring
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
        labels:
          service: 'prometheus'

  # Go-Orchestrator metrics
  - job_name: 'go-orchestrator'
    static_configs:
      - targets: ['go-orchestrator:9091']
        labels:
          service: 'go-orchestrator'
          environment: 'development'

  # Drone-Service metrics
  - job_name: 'drone-service'
    static_configs:
      - targets: ['drone-service:9092']
        labels:
          service: 'drone-service'
          environment: 'development'

  # RabbitMQ metrics
  - job_name: 'rabbitmq'
    static_configs:
      - targets: ['rabbitmq:15692']
        labels:
          service: 'rabbitmq'
          environment: 'development'

  # Container metrics
  - job_name: 'cadvisor'
    static_configs:
      - targets: ['cadvisor:8080']
        labels:
          service: 'cadvisor'
          environment: 'development'

  # Host machine metrics
  - job_name: 'node-exporter'
    static_configs:
      - targets: ['node-exporter:9100']
        labels:
          service: 'node-exporter'
          environment: 'development'

  # PostgreSQL metrics
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
        labels:
          service: 'postgres'
          environment: 'development'
```

**Configuration Parameters**:
- `scrape_interval`: How often to scrape targets (default: 15s)
- `evaluation_interval`: How often to evaluate alert rules (default: 15s)
- `targets`: Service endpoints to scrape metrics from
- `labels`: Metadata attached to all metrics from target

---

### Metrics Collection

#### Metrics Endpoints

| Service | Endpoint | Port | Metrics Exposed |
|---------|----------|------|-----------------|
| Go-Orchestrator | `/metrics` | 9091 | HTTP, gRPC, custom business metrics |
| Drone-Service | `/metrics` | 9092 | HTTP, WebSocket, drone telemetry |
| RabbitMQ | `/metrics` | 15692 | Queue, connection, consumer metrics |
| PostgreSQL | `/` | 9187 | Database, query, connection pool metrics |
| cAdvisor | `/metrics` | 8080 | Container CPU, memory, network, disk |
| Node Exporter | `/metrics` | 9100 | Host CPU, memory, disk, network |

#### Metric Types

**1. Counter**: Monotonically increasing value (e.g., total requests)
```promql
http_requests_total{service="go-orchestrator"}
```

**2. Gauge**: Value that can go up or down (e.g., active connections)
```promql
active_websocket_connections{service="drone-service"}
```

**3. Histogram**: Distribution of values (e.g., request duration)
```promql
http_request_duration_seconds_bucket{service="go-orchestrator"}
```

**4. Summary**: Similar to histogram but calculates quantiles on client side
```promql
grpc_request_duration_seconds{quantile="0.95"}
```

---

### Custom Metrics

#### Go-Orchestrator Metrics

**Implementation**: `backend/go-orchestrator/internal/controller/grpc/middleware/prometheus.go`

**HTTP Metrics** (via `go-gin-prometheus`):
```go
// Automatic metrics from gin-prometheus middleware
http_requests_total{method, endpoint, status}
http_request_duration_seconds{method, endpoint}
http_request_size_bytes{method, endpoint}
http_response_size_bytes{method, endpoint}
```

**gRPC Metrics** (custom implementation):
```go
var (
    grpcRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total",
            Help: "Total number of gRPC requests",
        },
        []string{"method", "code"},
    )

    grpcRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "grpc_request_duration_seconds",
            Help:    "gRPC request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method"},
    )
)
```

**Usage Example**:
```go
func PrometheusUnaryInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
        start := time.Now()
        resp, err := handler(ctx, req)
        duration := time.Since(start).Seconds()
        
        code := "OK"
        if err != nil {
            code = "ERROR"
        }

        grpcRequestsTotal.WithLabelValues(info.FullMethod, code).Inc()
        grpcRequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

        return resp, err
    }
}
```

**Business Metrics** (recommended additions):
```go
// Orders
ordersCreated := promauto.NewCounter(prometheus.CounterOpts{
    Name: "orders_created_total",
    Help: "Total orders created",
})

ordersCompleted := promauto.NewCounter(prometheus.CounterOpts{
    Name: "orders_completed_total",
    Help: "Total orders completed",
})

// Deliveries
activeDeliveries := promauto.NewGauge(prometheus.GaugeOpts{
    Name: "active_deliveries",
    Help: "Current number of active deliveries",
})

deliveryDuration := promauto.NewHistogram(prometheus.HistogramOpts{
    Name:    "delivery_duration_seconds",
    Help:    "Time from order creation to completion",
    Buckets: []float64{60, 300, 600, 1200, 1800, 3600},
})
```

---

#### Drone-Service Metrics

**Implementation**: `backend/drone-service/internal/app/app.go`

**HTTP/WebSocket Metrics** (via `go-gin-prometheus`):
```go
prometheusMiddleware := ginprometheus.NewPrometheus("drone-service")
prometheusMiddleware.Use(router)
```

**Custom Drone Metrics** (recommended):
```go
// Drone fleet status
dronesActive := promauto.NewGauge(prometheus.GaugeOpts{
    Name: "drones_active",
    Help: "Number of active drones",
})

dronesBatteryLevel := promauto.NewGaugeVec(prometheus.GaugeOpts{
    Name: "drone_battery_level",
    Help: "Battery level per drone",
}, []string{"drone_id"})

// WebSocket connections
wsConnections := promauto.NewGaugeVec(prometheus.GaugeOpts{
    Name: "websocket_connections",
    Help: "Active WebSocket connections",
}, []string{"connection_type"}) // connection_type: drone, admin, video

// Video streaming
videoFramesReceived := promauto.NewCounterVec(prometheus.CounterOpts{
    Name: "video_frames_received_total",
    Help: "Total video frames received",
}, []string{"drone_id"})
```

---

#### RabbitMQ Metrics

**Built-in Prometheus Plugin**: Enabled by default

**Key Metrics**:
```promql
# Queue metrics
rabbitmq_queue_messages{queue}           # Messages in queue
rabbitmq_queue_messages_ready{queue}     # Ready to consume
rabbitmq_queue_messages_unacked{queue}   # Unacknowledged
rabbitmq_queue_consumers{queue}          # Active consumers

# Connection metrics
rabbitmq_connections                     # Total connections
rabbitmq_channels                        # Total channels

# Node metrics
rabbitmq_node_mem_used                   # Memory usage
rabbitmq_node_disk_free                  # Free disk space
rabbitmq_node_fd_used                    # File descriptors used
```

---

#### PostgreSQL Metrics

**Exporter**: `prometheuscommunity/postgres-exporter`

**Connection String**:
```
postgresql://user:password@postgres:5432/skypost_delivery?sslmode=disable
```

**Key Metrics**:
```promql
# Database status
pg_up                                    # Database is up (1) or down (0)

# Connections
pg_stat_database_numbackends             # Active connections
pg_settings_max_connections              # Max connection limit

# Query performance
pg_stat_database_xact_commit             # Committed transactions
pg_stat_database_xact_rollback           # Rolled back transactions
pg_stat_database_deadlocks               # Deadlock count

# Cache performance
pg_stat_database_blks_hit                # Buffer cache hits
pg_stat_database_blks_read               # Disk reads

# Table statistics
pg_stat_user_tables_seq_scan             # Sequential scans
pg_stat_user_tables_idx_scan             # Index scans
pg_stat_user_tables_n_tup_ins            # Rows inserted
pg_stat_user_tables_n_tup_upd            # Rows updated
pg_stat_user_tables_n_tup_del            # Rows deleted
```

---

#### Container Metrics (cAdvisor)

**Metrics Collected**:
```promql
# CPU
container_cpu_usage_seconds_total        # Total CPU usage
container_cpu_system_seconds_total       # System CPU time
container_cpu_user_seconds_total         # User CPU time

# Memory
container_memory_usage_bytes             # Current memory usage
container_memory_max_usage_bytes         # Max memory usage
container_spec_memory_limit_bytes        # Memory limit

# Network
container_network_receive_bytes_total    # Bytes received
container_network_transmit_bytes_total   # Bytes transmitted
container_network_receive_errors_total   # Receive errors
container_network_transmit_errors_total  # Transmit errors

# Disk I/O
container_fs_reads_bytes_total           # Bytes read from disk
container_fs_writes_bytes_total          # Bytes written to disk
```

---

#### Host Metrics (Node Exporter)

**Metrics Collected**:
```promql
# CPU
node_cpu_seconds_total{mode}             # CPU time by mode (idle, user, system)
node_load1                               # 1-minute load average
node_load5                               # 5-minute load average
node_load15                              # 15-minute load average

# Memory
node_memory_MemTotal_bytes               # Total memory
node_memory_MemAvailable_bytes           # Available memory
node_memory_Buffers_bytes                # Buffer cache
node_memory_Cached_bytes                 # Page cache

# Disk
node_filesystem_size_bytes               # Filesystem size
node_filesystem_avail_bytes              # Available space
node_disk_read_bytes_total               # Bytes read
node_disk_written_bytes_total            # Bytes written

# Network
node_network_receive_bytes_total         # Bytes received
node_network_transmit_bytes_total        # Bytes transmitted
```

---

## Grafana

### Grafana Dashboards

**Dashboard Organization**:

```
Grafana Dashboards
├── Go Orchestrator/
│   ├── metrics.json       (HTTP, gRPC, business metrics)
│   └── logs.json          (Application logs)
├── Drone Service/
│   ├── metrics.json       (WebSocket, drone telemetry)
│   └── logs.json          (Application logs)
├── Admin Panel/
│   ├── metrics.json       (Frontend performance)
│   └── logs.json          (Access logs)
├── System/
│   ├── metrics.json       (Host CPU, memory, disk, network)
│   └── logs.json          (System logs)
├── RabbitMQ/
│   ├── metrics.json       (Queue, connection metrics)
│   └── logs.json          (RabbitMQ logs)
└── PostgreSQL/
    └── metrics.json       (Database performance)
```

---

#### Go-Orchestrator Dashboard

**File**: `monitoring/grafana/dashboards/go-orchestrator/metrics.json`

**Panels**:

1. **HTTP Request Rate**
   - Query: `rate(http_requests_total{service="go-orchestrator"}[5m])`
   - Visualization: Time series graph
   - Legend: `{{method}} {{endpoint}} - {{status}}`

2. **HTTP Request Duration (p95)**
   - Query: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{service="go-orchestrator"}[5m]))`
   - Visualization: Time series graph
   - Unit: Seconds

3. **HTTP Error Rate (5xx)**
   - Query: `rate(http_requests_total{service="go-orchestrator",status=~"5.."}[5m])`
   - Visualization: Time series graph (red)
   - Alert threshold: > 0.1 req/s

4. **gRPC Request Rate**
   - Query: `rate(grpc_requests_total{service="go-orchestrator"}[5m])`
   - Legend: `{{method}} - {{code}}`

5. **gRPC Request Duration (p95)**
   - Query: `histogram_quantile(0.95, rate(grpc_request_duration_seconds_bucket{service="go-orchestrator"}[5m]))`

6. **Active Database Connections**
   - Query: `pg_stat_database_numbackends{datname="skypost_delivery"}`

7. **Memory Usage**
   - Query: `container_memory_usage_bytes{name=~".*go-orchestrator.*"} / container_spec_memory_limit_bytes{name=~".*go-orchestrator.*"}`
   - Unit: Percentage

8. **CPU Usage**
   - Query: `rate(container_cpu_usage_seconds_total{name=~".*go-orchestrator.*"}[5m])`
   - Unit: Percentage

---

#### Drone-Service Dashboard

**File**: `monitoring/grafana/dashboards/drone-service/metrics.json`

**Panels**:

1. **Active Drone Connections**
   - Query: `websocket_connections{connection_type="drone"}`
   - Visualization: Stat panel

2. **Active Admin Connections**
   - Query: `websocket_connections{connection_type="admin"}`
   - Visualization: Stat panel

3. **Drone Battery Levels**
   - Query: `drone_battery_level`
   - Visualization: Gauge (per drone)
   - Thresholds: Red < 20%, Yellow < 50%, Green >= 50%

4. **Active Deliveries**
   - Query: `active_deliveries`
   - Visualization: Stat panel

5. **Video Frame Rate**
   - Query: `rate(video_frames_received_total[1m])`
   - Unit: Frames per second
   - Legend: `Drone {{drone_id}}`

6. **WebSocket Message Rate**
   - Query: `rate(websocket_messages_total[5m])`
   - Legend: `{{message_type}}`

---

#### System Dashboard

**File**: `monitoring/grafana/dashboards/system/metrics.json`

**Panels**:

1. **CPU Usage**
   - Query: `100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
   - Unit: Percentage

2. **Memory Usage**
   - Query: `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
   - Unit: Percentage

3. **Disk Usage**
   - Query: `(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100`
   - Unit: Percentage
   - Legend: `{{mountpoint}}`

4. **Network Traffic**
   - Queries:
     - Ingress: `rate(node_network_receive_bytes_total[5m])`
     - Egress: `rate(node_network_transmit_bytes_total[5m])`
   - Unit: Bytes/sec

5. **Load Average**
   - Queries:
     - 1m: `node_load1`
     - 5m: `node_load5`
     - 15m: `node_load15`

6. **Disk I/O**
   - Queries:
     - Read: `rate(node_disk_read_bytes_total[5m])`
     - Write: `rate(node_disk_written_bytes_total[5m])`
   - Unit: Bytes/sec

---

#### RabbitMQ Dashboard

**File**: `monitoring/grafana/dashboards/rabbitmq/metrics.json`

**Panels**:

1. **Queue Message Count**
   - Query: `rabbitmq_queue_messages`
   - Legend: `{{queue}}`

2. **Message Rate**
   - Queries:
     - Publish: `rate(rabbitmq_queue_messages_published_total[5m])`
     - Consume: `rate(rabbitmq_queue_messages_consumed_total[5m])`
   - Legend: `{{queue}}`

3. **Consumer Count**
   - Query: `rabbitmq_queue_consumers`
   - Legend: `{{queue}}`
   - Alert: = 0 (no consumers)

4. **Memory Usage**
   - Query: `rabbitmq_node_mem_used / rabbitmq_node_mem_limit * 100`
   - Unit: Percentage
   - Threshold: 90% warning

5. **Connection Count**
   - Query: `rabbitmq_connections`
   - Visualization: Stat panel

6. **Disk Space**
   - Query: `rabbitmq_node_disk_free`
   - Unit: Bytes
   - Threshold: < 1GB warning

---

#### PostgreSQL Dashboard

**File**: `monitoring/grafana/dashboards/postgres/metrics.json`

**Panels**:

1. **Database Status**
   - Query: `pg_up`
   - Visualization: Stat (1 = up, 0 = down)

2. **Active Connections**
   - Query: `pg_stat_database_numbackends{datname="skypost_delivery"}`
   - Threshold: 80% of `pg_settings_max_connections`

3. **Transaction Rate**
   - Queries:
     - Commits: `rate(pg_stat_database_xact_commit[5m])`
     - Rollbacks: `rate(pg_stat_database_xact_rollback[5m])`

4. **Cache Hit Ratio**
   - Query: `pg_stat_database_blks_hit / (pg_stat_database_blks_hit + pg_stat_database_blks_read) * 100`
   - Unit: Percentage
   - Target: > 90%

5. **Deadlocks**
   - Query: `rate(pg_stat_database_deadlocks[5m])`
   - Alert: > 0

6. **Table Operations**
   - Queries:
     - Inserts: `rate(pg_stat_user_tables_n_tup_ins[5m])`
     - Updates: `rate(pg_stat_user_tables_n_tup_upd[5m])`
     - Deletes: `rate(pg_stat_user_tables_n_tup_del[5m])`

7. **Query Duration (Slowest)**
   - Query: `pg_stat_activity_max_tx_duration`
   - Unit: Seconds
   - Alert: > 300s (5 minutes)

---

### Data Sources

**Configuration File**: `monitoring/grafana/provisioning/datasources/datasources.yml`

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    uid: prometheus
    url: http://prometheus:9090
    isDefault: true
    editable: true
    jsonData:
      timeInterval: 15s
      httpMethod: POST

  - name: Loki
    type: loki
    access: proxy
    uid: loki
    url: http://loki:3100
    editable: true
    jsonData:
      maxLines: 1000
      derivedFields:
        - datasourceUid: prometheus
          matcherRegex: "traceID=(\\w+)"
          name: TraceID
          url: "$${__value.raw}"
```

**Data Source Parameters**:
- `access: proxy`: Grafana server proxies requests (recommended for security)
- `timeInterval`: Minimum interval between data points
- `httpMethod: POST`: Use POST for large queries (prevents URL length limits)
- `maxLines`: Maximum log lines to return from Loki

---

### Visualization

**Panel Types**:

1. **Time Series Graph**
   - Use for: Trends over time (CPU, memory, request rate)
   - Options: Line, area, bars

2. **Stat**
   - Use for: Single value metrics (uptime, connection count)
   - Options: Number, gauge, sparkline

3. **Gauge**
   - Use for: Percentage metrics (0-100%)
   - Thresholds: Green (good), yellow (warning), red (critical)

4. **Table**
   - Use for: Multiple series with many labels
   - Sort, filter, and format columns

5. **Heatmap**
   - Use for: Distribution over time (request duration buckets)

6. **Logs Panel**
   - Use for: Log exploration from Loki
   - Live streaming, filtering, highlighting

**Refresh Intervals**:
- Real-time monitoring: 5s
- General dashboards: 10s
- Historical analysis: 1m+

**Time Ranges**:
- Last 5 minutes: Troubleshooting
- Last 1 hour: Current operations
- Last 24 hours: Daily patterns
- Last 7 days: Weekly trends

---

## Loki & Promtail

### Log Aggregation

**Loki Configuration**: `monitoring/loki/loki-config.yml`

```yaml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1

schema_config:
  configs:
    - from: 2020-10-24
      store: tsdb              # Time series database
      object_store: filesystem # Local filesystem storage
      schema: v13
      index:
        prefix: index_
        period: 24h            # Daily indexes

limits_config:
  reject_old_samples: true
  reject_old_samples_max_age: 168h  # 7 days

compactor:
  working_directory: /loki/compactor
  compaction_interval: 10m
  retention_enabled: false
```

**Storage Structure**:
```
/loki/
├── chunks/          # Log data stored as chunks
├── rules/           # Alert rules (future)
└── compactor/       # Compaction working directory
```

---

**Promtail Configuration**: `monitoring/promtail/promtail-config.yml`

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml  # Track read positions

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  # Go-Orchestrator logs
  - job_name: go-orchestrator
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
        filters:
          - name: label
            values: ["com.docker.compose.service=go-orchestrator"]
    relabel_configs:
      - source_labels: ['__meta_docker_container_name']
        regex: '/(.*)'
        target_label: 'container_name'
      - source_labels: ['__meta_docker_container_label_com_docker_compose_service']
        target_label: 'service'

  # Similar configs for:
  # - drone-service
  # - admin-panel
  # - postgres
  # - minio
```

**Log Collection Process**:
1. Promtail discovers Docker containers via socket
2. Reads container logs from `/var/lib/docker/containers/`
3. Extracts metadata (container name, service label)
4. Forwards logs to Loki with labels
5. Tracks read position to avoid duplicates

---

### Log Queries

**LogQL Syntax** (Loki Query Language):

#### Basic Queries

**1. Filter by service**:
```logql
{service="go-orchestrator"}
```

**2. Filter by container name**:
```logql
{container_name="skypost-delivery-go-orchestrator-dev"}
```

**3. Search for text**:
```logql
{service="go-orchestrator"} |= "error"
```

**4. Exclude text**:
```logql
{service="go-orchestrator"} != "debug"
```

**5. Regex filter**:
```logql
{service="go-orchestrator"} |~ "error|fatal|panic"
```

---

#### Advanced Queries

**1. Parse JSON logs**:
```logql
{service="go-orchestrator"} | json | level="error"
```

**2. Count errors per minute**:
```logql
sum(rate({service="go-orchestrator"} |= "error" [1m]))
```

**3. Top 10 error messages**:
```logql
topk(10, sum by (msg) (rate({service="go-orchestrator"} | json | level="error" [5m])))
```

**4. HTTP 5xx errors with endpoint**:
```logql
{service="go-orchestrator"} | json | status >= 500 | line_format "{{.method}} {{.path}} - {{.status}}"
```

**5. Slow database queries**:
```logql
{service="postgres"} |~ "duration: [0-9]{4,} ms"
```

**6. Drone disconnections**:
```logql
{service="drone-service"} |~ "drone .* disconnected"
```

---

#### Log Aggregations

**1. Error rate by service**:
```logql
sum(rate({job=~".+"} |= "error" [5m])) by (service)
```

**2. Request count by endpoint**:
```logql
sum(rate({service="go-orchestrator"} | json | __error__="" [5m])) by (path)
```

**3. Average response time**:
```logql
avg_over_time({service="go-orchestrator"} | json | unwrap duration [5m])
```

---

## Alerting

### Alert Rules

**Configuration File**: `monitoring/prometheus/alerts.yml`

#### Service Alerts

**1. Service Down**
```yaml
- alert: ServiceDown
  expr: up == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Service {{ $labels.job }} is down"
    description: "Service {{ $labels.job }} has been down for more than 1 minute."
```

**2. High Error Rate**
```yaml
- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High error rate for {{ $labels.job }}"
    description: "Error rate is above 10% for {{ $labels.job }}"
```

**3. High Memory Usage**
```yaml
- alert: HighMemoryUsage
  expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High memory usage for {{ $labels.name }}"
    description: "Memory usage is above 90% for container {{ $labels.name }}"
```

**4. High CPU Usage**
```yaml
- alert: HighCPUUsage
  expr: rate(container_cpu_usage_seconds_total[5m]) > 0.8
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High CPU usage for {{ $labels.name }}"
    description: "CPU usage is above 80% for container {{ $labels.name }}"
```

---

#### RabbitMQ Alerts

**1. Queue Too Long**
```yaml
- alert: RabbitMQQueueTooLong
  expr: rabbitmq_queue_messages > 1000
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "RabbitMQ queue {{ $labels.queue }} is too long"
    description: "Queue {{ $labels.queue }} has {{ $value }} messages"
```

**2. No Consumers**
```yaml
- alert: RabbitMQNoConsumers
  expr: rabbitmq_queue_consumers == 0
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "RabbitMQ queue {{ $labels.queue }} has no consumers"
    description: "No consumers connected to queue {{ $labels.queue }}"
```

**3. High Memory Usage**
```yaml
- alert: RabbitMQHighMemoryUsage
  expr: rabbitmq_node_mem_used / rabbitmq_node_mem_limit > 0.9
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "RabbitMQ memory usage is high"
    description: "RabbitMQ is using {{ $value | humanizePercentage }} of available memory"
```

**4. Disk Space Low**
```yaml
- alert: RabbitMQDiskSpaceLow
  expr: rabbitmq_node_disk_free < 1000000000  # < 1GB
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "RabbitMQ disk space is low"
    description: "RabbitMQ has less than 1GB of free disk space"
```

---

#### PostgreSQL Alerts

**1. Database Down**
```yaml
- alert: PostgreSQLDown
  expr: pg_up == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "PostgreSQL is down"
    description: "PostgreSQL database is not responding"
```

**2. Too Many Connections**
```yaml
- alert: PostgreSQLTooManyConnections
  expr: pg_stat_database_numbackends{datname="skypost_delivery"} / pg_settings_max_connections > 0.8
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL has too many connections"
    description: "PostgreSQL is using {{ $value | humanizePercentage }} of max connections"
```

**3. Deadlocks Detected**
```yaml
- alert: PostgreSQLDeadlocks
  expr: rate(pg_stat_database_deadlocks{datname="skypost_delivery"}[5m]) > 0
  for: 1m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL deadlocks detected"
    description: "Deadlocks detected in database {{ $labels.datname }}"
```

**4. Slow Queries**
```yaml
- alert: PostgreSQLSlowQueries
  expr: pg_stat_activity_max_tx_duration > 300  # > 5 minutes
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL has slow queries"
    description: "Query running for more than 5 minutes detected"
```

**5. Low Cache Hit Ratio**
```yaml
- alert: PostgreSQLLowCacheHitRatio
  expr: (pg_stat_database_blks_hit / (pg_stat_database_blks_hit + pg_stat_database_blks_read)) < 0.9
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL cache hit ratio is low"
    description: "Cache hit ratio is {{ $value | humanizePercentage }}"
```

---

### Alert Channels

**Supported Notification Channels** (via Grafana):

1. **Email**
2. **Slack**
3. **Telegram**
4. **PagerDuty**
5. **Webhook** (custom integrations)

**Configuration** (Grafana UI):
1. Navigate to **Alerting > Contact Points**
2. Click **New contact point**
3. Choose notification type
4. Configure credentials/webhook URL
5. Test notification
6. Create notification policy to route alerts

**Recommended Setup**:
- **Critical alerts** (ServiceDown, PostgreSQLDown): PagerDuty + Slack
- **Warning alerts** (HighMemoryUsage, SlowQueries): Slack
- **Info alerts**: Email digest

---

## Exporters

### Exporter Summary

| Exporter | Port | Purpose | Metrics |
|----------|------|---------|---------|
| Go-Orchestrator | 9091 | HTTP/gRPC service metrics | 50+ |
| Drone-Service | 9092 | WebSocket/drone metrics | 40+ |
| RabbitMQ | 15692 | Message queue metrics | 200+ |
| PostgreSQL Exporter | 9187 | Database metrics | 100+ |
| cAdvisor | 8080 | Container metrics | 80+ |
| Node Exporter | 9100 | Host machine metrics | 300+ |

---

### Container-Specific Configuration

**Docker Compose Labels**:
```yaml
services:
  go-orchestrator:
    labels:
      - "logging=promtail"
      - "logging_jobname=containerlogs"
```

**Purpose**:
- `logging=promtail`: Marks container for log collection
- `logging_jobname`: Groups logs under specific job name

---

## Container Monitoring

### cAdvisor

**Purpose**: Collects container resource usage and performance metrics

**Access**: `http://localhost:8080` (Web UI)

**Metrics Collected**:
- CPU usage per container
- Memory usage and limits
- Network I/O
- Disk I/O
- Process count

**Docker Compose Config**:
```yaml
cadvisor:
  image: gcr.io/cadvisor/cadvisor:latest
  privileged: true
  devices:
    - /dev/kmsg:/dev/kmsg
  volumes:
    - /:/rootfs:ro
    - /var/run:/var/run:ro
    - /sys:/sys:ro
    - /var/lib/docker/:/var/lib/docker:ro
    - /dev/disk/:/dev/disk:ro
    - /etc/machine-id:/etc/machine-id:ro
```

**Key Queries**:
```promql
# Top 5 containers by memory
topk(5, container_memory_usage_bytes)

# Container CPU usage percentage
rate(container_cpu_usage_seconds_total[5m]) * 100

# Network bandwidth
rate(container_network_transmit_bytes_total[5m])
```

---

### Node Exporter

**Purpose**: Collects host machine metrics

**Access**: Metrics only (no Web UI)

**Metrics Collected**:
- CPU usage per core
- Memory usage (total, available, buffers, cache)
- Disk usage and I/O
- Network traffic and errors
- Load average
- File descriptor usage

**Docker Compose Config**:
```yaml
node-exporter:
  image: prom/node-exporter:latest
  command:
    - '--path.procfs=/host/proc'
    - '--path.rootfs=/rootfs'
    - '--path.sysfs=/host/sys'
    - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
  volumes:
    - /proc:/host/proc:ro
    - /sys:/host/sys:ro
    - /:/rootfs:ro
```

---

## Best Practices

### Metrics Best Practices

1. **Use Consistent Naming**:
   - Format: `<namespace>_<subsystem>_<name>_<unit>`
   - Example: `http_request_duration_seconds`

2. **Choose Appropriate Metric Types**:
   - Counter: Monotonically increasing (requests, errors)
   - Gauge: Can go up/down (temperature, active connections)
   - Histogram: Distribution (request duration, response size)

3. **Use Labels Wisely**:
   - High cardinality kills performance (avoid user IDs, timestamps)
   - Good: `method`, `status`, `endpoint`
   - Bad: `user_id`, `request_id`

4. **Set Reasonable Scrape Intervals**:
   - 15s default (good for most use cases)
   - 5s for critical real-time metrics
   - 1m for slow-changing metrics

5. **Avoid Excessive Labels**:
   - Limit: 5-10 labels per metric
   - Each unique label combination = new time series

---

### Alerting Best Practices

1. **Set Appropriate Thresholds**:
   - Based on historical data
   - Leave room for spikes (don't alert on every anomaly)
   - Use `for` clause to avoid flapping

2. **Use Alert Severity Levels**:
   - **Critical**: Requires immediate action (service down)
   - **Warning**: Attention needed (high memory, slow queries)
   - **Info**: Informational (version upgrade)

3. **Write Actionable Alerts**:
   - Include what's wrong
   - Include affected service/component
   - Include runbook link (future)

4. **Avoid Alert Fatigue**:
   - Don't alert on expected behavior
   - Aggregate similar alerts
   - Use silences during maintenance

---

### Dashboard Best Practices

1. **Organize Dashboards by Audience**:
   - Executive: High-level KPIs
   - DevOps: System health, resource usage
   - Developers: Application metrics, errors

2. **Use Consistent Time Ranges**:
   - Allow users to adjust via dropdown
   - Default: Last 1 hour

3. **Show Context**:
   - Include related metrics on same panel
   - Use annotations for deployments

4. **Optimize Query Performance**:
   - Use recording rules for expensive queries
   - Limit time range for large datasets

---

### Log Management Best Practices

1. **Structure Logs**:
   - Use JSON format for easy parsing
   - Include: timestamp, level, message, context

2. **Use Appropriate Log Levels**:
   - ERROR: Application errors requiring investigation
   - WARN: Potential issues (retries, deprecated APIs)
   - INFO: Important business events (order created)
   - DEBUG: Detailed debugging information

3. **Add Context**:
   - Include request ID for tracing
   - Include user ID (anonymized if needed)
   - Include relevant business data

4. **Set Retention Policies**:
   - Development: 7 days
   - Production: 30+ days
   - Compliance: Based on regulations

---

## Troubleshooting

### Common Issues

#### 1. Prometheus Not Scraping Targets

**Symptoms**:
- Target shows as "DOWN" in Prometheus UI
- No metrics in Grafana

**Diagnosis**:
```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Check service endpoint directly
curl http://go-orchestrator:9091/metrics
```

**Solutions**:
- Verify service is running: `docker ps`
- Check network connectivity: Services must be on same Docker network
- Verify metrics endpoint is exposed
- Check firewall rules

---

#### 2. High Cardinality Metrics

**Symptoms**:
- Prometheus memory usage increasing
- Slow query performance
- "Too many samples" errors

**Diagnosis**:
```promql
# Check time series count
count({__name__=~".+"})

# Top metrics by cardinality
topk(10, count by (__name__)({__name__=~".+"}))
```

**Solutions**:
- Remove high-cardinality labels (user IDs, request IDs)
- Use relabeling to drop unnecessary labels
- Increase Prometheus memory limit

---

#### 3. Loki Log Ingestion Issues

**Symptoms**:
- No logs in Grafana
- Promtail errors in logs

**Diagnosis**:
```bash
# Check Promtail status
docker logs skypost-delivery-promtail-dev

# Check Loki ingestion rate
curl http://localhost:3100/metrics | grep loki_ingester
```

**Solutions**:
- Verify Promtail can access Docker socket: `/var/run/docker.sock`
- Check Promtail positions file: `/tmp/positions.yaml`
- Ensure Loki is reachable: `curl http://loki:3100/ready`
- Verify container labels match Promtail config

---

#### 4. Grafana Dashboard Not Loading Data

**Symptoms**:
- "No data" message in panels
- Query errors in panel

**Diagnosis**:
- Check data source connection: **Configuration > Data Sources**
- Test query in Explore: **Explore > Prometheus/Loki**
- Check query syntax

**Solutions**:
- Verify Prometheus/Loki URLs in data source config
- Check time range (ensure data exists for selected period)
- Simplify query to isolate issue
- Check Prometheus/Loki logs for errors

---

#### 5. Alert Not Firing

**Symptoms**:
- Expected alert not triggered
- No notification received

**Diagnosis**:
```bash
# Check alert state in Prometheus
curl http://localhost:9090/api/v1/alerts

# Check alert rule configuration
curl http://localhost:9090/api/v1/rules
```

**Solutions**:
- Verify alert expression returns data
- Check `for` duration (alert fires after threshold period)
- Verify alert rule loaded: Check Prometheus logs
- Test notification channel in Grafana
- Check alert routing rules

---

### Performance Optimization

**Prometheus**:
```yaml
# Increase memory for high-load systems
command:
  - '--storage.tsdb.retention.time=30d'
  - '--storage.tsdb.retention.size=50GB'
  - '--storage.tsdb.path=/prometheus'
```

**Loki**:
```yaml
# Increase ingestion limits
limits_config:
  ingestion_rate_mb: 10
  ingestion_burst_size_mb: 20
  per_stream_rate_limit: 3MB
  per_stream_rate_limit_burst: 15MB
```

**Grafana**:
```yaml
# Optimize query caching
GF_CACHING_BACKEND: redis
GF_CACHING_CONNSTR: addr=redis:6379
```

---

**Documentation Version**: 1.0  
**Last Updated**: December 13, 2025  
**Maintainers**: SkyPost Development Team
