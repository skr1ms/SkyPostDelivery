# End-to-End Delivery Scenario

This document walks through a full delivery lifecycle: a customer places an order in the mobile app, the system schedules the mission, the drone executes the flight, the parcel is delivered, and the drone returns to the base marker `131`.

## 1. Customer Places an Order (Mobile App)
1. User opens the Flutter mobile app and authenticates (`POST /api/v1/auth/login`).
2. In **Goods** catalogue, the user selects an item and confirms purchase.
3. Mobile app calls orchestrator: `POST /api/v1/orders` with `{ "good_id": <UUID> }`.
4. Orchestrator `OrderUseCase.CreateOrder`:
   - Validates stock (decrements quantity in `good_repo`).
   - Reserves a compatible locker cell (`lockerRepo.FindAvailableCell` → status `reserved`).
   - Selects an available parcel automat (closest/first working).
   - Attempts to allocate an idle drone (`droneRepo.GetAvailable`).
5. Database side effects: order row created (`status=pending`), delivery record created linking drone and automat (`status=pending`).

## 2. Mission Scheduling (Orchestrator → Drone-Service)
1. If a drone is available, orchestrator publishes a `DeliveryTask` message to RabbitMQ queue `deliveries`.
   - Payload includes `delivery_id`, `order_id`, automat details, ArUco marker, parcel dimensions.
2. The Python `drone-service` runs `DeliveryWorker` (`app/services/workers/delivery.py`) consuming `deliveries` queue.
3. `DeliveryUseCase.execute_delivery` persists task state, ensures drone is registered, and notifies the WebSocket handler to send the task to the drone.
4. Drone connects via WebSocket (`/ws/drone`), registers by IP, receives task payload.

## 3. Drone Takeoff & Transit (Clover ROS)
1. Drone-side parser `delivery_service.execute_delivery` (ROS script) initialises Clover API and navigation controller.
2. Sequence:
   - Takeoff from current pad (marker `131` by default) to cruise altitude.
   - Navigate to target ArUco marker provided in task (`navigate_to_marker`).
   - Transition to `DELIVERING` state and perform `land_on_marker`.
3. Status updates: drone sends WebSocket messages (`status_update`, `delivery_update`) back to drone-service; drone-service relays to orchestrator/admin dashboards.

## 4. Arrival & Parcel Drop
1. Upon landing, drone emits `arrived_at_destination` update containing `order_id`, `parcel_automat_id`.
2. Drone-service triggers orchestrator gRPC `RequestCellOpen` to open the reserved locker cell.
3. Once the operator/orchestrator authorises cargo drop (`drop_cargo` command published over WebSocket), drone-side script:
   - Waits for `drop_cargo` signal (`cargo_drop_approved` flag).
   - Performs mechanical drop (`delivery_service` simulates 3-second wait).
   - Sends `cargo_dropped` update with `drone_status="cargo_dropped"`.
4. Drone-service notifies orchestrator to mark delivery complete (`DeliveryStatus.COMPLETED`), locker cell becomes `available`, order status transitions to `completed`.

## 5. Drone Return to Base (`return_to_base`)
1. After drop, drone-side logic transitions to finalisation:
   - Drone-service sends `return_to_base` command specifying base marker `131`.
   - ROS navigation executes takeoff if needed, flies to marker `131`, and lands.
2. Drone status set to `IDLE`; `drone_repo.UpdateStatus` in orchestrator updates to available.
3. Drone-service acknowledges queue completion so the next task may be assigned.

## 6. Customer Notification & App Flow
1. Orchestrator may notify customer via push/SMS once order status becomes `completed`.
2. Mobile app polls or listens via WebSocket to update UI.
3. Customer sees order card marked as delivered with locker details.

## 7. Order Return Scenario (Optional)
If the user cancels (`POST /api/v1/orders/{id}/return`) while status is `pending`/`in_progress`:
1. Orchestrator frees locker cell, restores stock, and publishes a high-priority `delivery.return` RabbitMQ message.
2. `DeliveryWorker.handle_return_task` issues `return_to_base` command to the drone (if already airborne).
3. Drone aborts mission, returns to marker `131`, order marked `cancelled`.

## Data Flows Summary
- **HTTP REST:** Mobile app ↔ orchestrator for orders/auth.
- **gRPC:** Orchestrator ↔ drone-service for cell operations and status updates.
- **RabbitMQ:** Task queues `deliveries`, `deliveries.priority`, `delivery.return` ensure asynchronous assignment.
- **WebSocket:** Drone (ROS) ↔ drone-service (task commands, telemetry) and admin dashboard subscriptions.

## Operational Monitoring Touchpoints
- Prometheus metrics:
  - `go-orchestrator` job tracks HTTP/gRPC throughput and latency.
  - `drone-service` job exposes FastAPI request and task processing metrics.
- Loki logs:
  - Observe mission logs via `{service="drone-service"}` and `{service="go-orchestrator"}`.
- Grafana dashboards visualise active deliveries, drone status, RabbitMQ queue depth.

## Conclusion
This scenario covers the normal delivery flow plus cancellation handling. For deeper API details, see `docs/en/API-SPEC.md`; for deployment considerations, refer to `docs/en/DEPLOYMENT.md` and `docs/en/OBSERVABILITY.md`.
