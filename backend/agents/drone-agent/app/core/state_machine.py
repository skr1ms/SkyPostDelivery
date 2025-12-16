import logging
from typing import Optional
from ..models.task import DeliveryTask, DeliveryState

logger = logging.getLogger(__name__)


class StateMachine:
    def __init__(self):
        self.current_task: Optional[DeliveryTask] = None
        self._sent_events = set()

    def set_task(self, task: DeliveryTask):
        self.current_task = task
        self.current_task.state = DeliveryState.PENDING
        logger.info(f"New delivery task set: {task.delivery_id}")

    def transition_to(self, new_state: DeliveryState):
        if self.current_task is None:
            logger.warning(f"Cannot transition to {new_state}: no active task")
            return

        old_state = self.current_task.state
        self.current_task.state = new_state
        logger.info(f"State transition: {old_state} -> {new_state}")

    def is_event_sent(self, event_id: str) -> bool:
        return event_id in self._sent_events

    def mark_event_sent(self, event_id: str):
        self._sent_events.add(event_id)
        logger.debug(f"Event marked as sent: {event_id}")

    def clear_task(self):
        if self.current_task:
            logger.info(f"Clearing task: {self.current_task.delivery_id}")
        self.current_task = None
        self._sent_events.clear()

    def get_state(self) -> Optional[DeliveryState]:
        return self.current_task.state if self.current_task else None
