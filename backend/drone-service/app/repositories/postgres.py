import asyncpg
from typing import Optional
from app.models.models import DroneState, DeliveryTask, DeliveryStatus
from app.models.ports import StateRepositoryPort
from config.config import settings


class PostgresStateRepository(StateRepositoryPort):
    def __init__(self):
        self.pool: Optional[asyncpg.Pool] = None
    
    async def connect(self):
        self.pool = await asyncpg.create_pool(
            settings.DATABASE_URL,
            min_size=2,
            max_size=10
        )
        print(f"PostgreSQL pool connected: {settings.DATABASE_URL}")
    
    async def close(self):
        if self.pool:
            await self.pool.close()
            print("PostgreSQL pool closed")
    
    async def get_drone_id_by_ip(self, ip_address: str) -> Optional[str]:
        if not self.pool:
            raise RuntimeError("Database pool not initialized. Call connect() first.")
        
        try:
            async with self.pool.acquire() as conn:
                row = await conn.fetchrow(
                    "SELECT id FROM drones WHERE ip_address = $1",
                    ip_address
                )
                if row:
                    return str(row['id'])
                return None
        except Exception as e:
            print(f"Error getting drone_id by IP {ip_address}: {e}")
            return None
    
    async def update_drone_battery(self, drone_id: str, battery_level: float) -> bool:
        if not self.pool:
            raise RuntimeError("Database pool not initialized. Call connect() first.")
        
        try:
            async with self.pool.acquire() as conn:
                await conn.execute(
                    "SELECT update_drone_battery($1::uuid, $2::decimal)",
                    drone_id,
                    battery_level
                )
                return True
        except Exception as e:
            print(f"Error updating battery for drone {drone_id}: {e}")
            return False
    
    async def save_drone_state(self, state: DroneState) -> bool:
        if not self.pool:
            raise RuntimeError("Database pool not initialized. Call connect() first.")
        
        try:
            async with self.pool.acquire() as conn:
                await conn.execute(
                    """
                    UPDATE drones
                    SET 
                        status = $2,
                        battery_level = $3,
                        updated_at = CURRENT_TIMESTAMP
                    WHERE id = $1::uuid
                    """,
                    state.drone_id,
                    state.status.value,
                    state.battery_level
                )
                return True
        except Exception as e:
            print(f"Error saving drone state for {state.drone_id}: {e}")
            return False
    
    async def get_drone_state(self, drone_id: str) -> Optional[DroneState]:
        if not self.pool:
            raise RuntimeError("Database pool not initialized. Call connect() first.")
        
        try:
            async with self.pool.acquire() as conn:
                row = await conn.fetchrow(
                    """
                    SELECT id, status, battery_level
                    FROM drones
                    WHERE id = $1::uuid
                    """,
                    drone_id
                )
                
                if not row:
                    return None
                
                return None
        except Exception as e:
            print(f"Error getting drone state for {drone_id}: {e}")
            return None
    
    async def save_delivery_task(self, task: DeliveryTask) -> bool:
        if not self.pool:
            raise RuntimeError("Database pool not initialized. Call connect() first.")
        
        try:
            async with self.pool.acquire() as conn:
                await conn.execute(
                    """
                    UPDATE deliveries
                    SET 
                        drone_id = $2::uuid,
                        status = $3,
                        started_at = COALESCE(started_at, CURRENT_TIMESTAMP)
                    WHERE id = $1::uuid
                    """,
                    task.delivery_id,
                    task.drone_id,
                    task.status.value
                )
                return True
        except Exception as e:
            print(f"Error saving delivery task {task.delivery_id}: {e}")
            return False
    
    async def get_delivery_task(self, delivery_id: str) -> Optional[DeliveryTask]:
        return None
    
    async def update_delivery_status(
        self,
        delivery_id: str,
        status: DeliveryStatus,
        error_message: Optional[str] = None
    ) -> bool:
        if not self.pool:
            raise RuntimeError("Database pool not initialized. Call connect() first.")
        
        try:
            async with self.pool.acquire() as conn:
                if status == DeliveryStatus.COMPLETED:
                    await conn.execute(
                        """
                        UPDATE deliveries
                        SET 
                            status = $2,
                            completed_at = CURRENT_TIMESTAMP
                        WHERE id = $1::uuid
                        """,
                        delivery_id,
                        status.value
                    )
                else:
                    await conn.execute(
                        """
                        UPDATE deliveries
                        SET status = $2
                        WHERE id = $1::uuid
                        """,
                        delivery_id,
                        status.value
                    )
                return True
        except Exception as e:
            print(f"Error updating delivery status for {delivery_id}: {e}")
            return False
