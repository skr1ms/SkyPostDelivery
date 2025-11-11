import json
import logging
from typing import Callable, Optional, Coroutine, Any
from aio_pika import connect_robust, Message, IncomingMessage
from aio_pika.abc import AbstractRobustConnection, AbstractChannel, AbstractQueue

logger = logging.getLogger(__name__)


class RabbitMQClient:
    def __init__(self, rabbitmq_url: str):
        self.rabbitmq_url = rabbitmq_url
        self.connection: Optional[AbstractRobustConnection] = None
        self.channel: Optional[AbstractChannel] = None
        self.queues: dict[str, AbstractQueue] = {}
        
    async def connect(self):
        try:
            self.connection = await connect_robust(self.rabbitmq_url)
            self.channel = await self.connection.channel()
            await self.channel.set_qos(prefetch_count=1)
            logger.info("Connected to RabbitMQ successfully")
        except Exception as e:
            logger.error(f"Failed to connect to RabbitMQ: {e}")
            raise
            
    async def declare_queue(self, queue_name: str, durable: bool = True):
        if not self.channel:
            raise RuntimeError("Channel not initialized. Call connect() first.")
        
        arguments = {}
        if queue_name in ["deliveries", "deliveries.priority"]:
            arguments = {
                "x-dead-letter-exchange": "",
                "x-dead-letter-routing-key": "deliveries.dlq",
                "x-message-ttl": 3600000
            }
            
        queue = await self.channel.declare_queue(
            queue_name,
            durable=durable,
            arguments=arguments
        )
        self.queues[queue_name] = queue
        logger.info(f"Declared queue: {queue_name}")
        return queue
        
    async def publish(self, queue_name: str, message: dict):
        if not self.channel:
            raise RuntimeError("Channel not initialized. Call connect() first.")
            
        message_body = json.dumps(message).encode()
        
        await self.channel.default_exchange.publish(
            Message(body=message_body, delivery_mode=2),
            routing_key=queue_name
        )
        logger.debug(f"Published message to {queue_name}: {message}")
        
    async def consume(
        self,
        queue_name: str,
        callback: Callable[[dict], Coroutine[Any, Any, None]]
    ):
        queue = self.queues.get(queue_name)
        if not queue:
            queue = await self.declare_queue(queue_name)
            
        async def process_message(message: IncomingMessage):
            async with message.process():
                try:
                    body = json.loads(message.body.decode())
                    logger.info(f"Received message from {queue_name}: {body}")
                    await callback(body)
                except Exception as e:
                    logger.error(f"Error processing message: {e}")
                    raise
                    
        await queue.consume(process_message)
        logger.info(f"Started consuming from queue: {queue_name}")
        
    async def close(self):
        if self.connection:
            await self.connection.close()
            logger.info("Closed RabbitMQ connection")

