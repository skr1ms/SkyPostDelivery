import pytest
from app.core.state_machine import StateMachine
from app.models.task import DeliveryTask, DeliveryState


def test_state_machine_initialization():
    sm = StateMachine()
    assert sm.current_task is None
    assert len(sm._sent_events) == 0


def test_set_task():
    sm = StateMachine()
    task = DeliveryTask(
        delivery_id="test-123",
        order_id="order-456",
        good_id="good-789",
        parcel_automat_id="automat-1",
        target_aruco_id=135,
        home_aruco_id=131,
    )
    
    sm.set_task(task)
    
    assert sm.current_task == task
    assert sm.current_task.state == DeliveryState.PENDING


def test_transition_to():
    sm = StateMachine()
    task = DeliveryTask(
        delivery_id="test-123",
        order_id="order-456",
        good_id="good-789",
        parcel_automat_id="automat-1",
        target_aruco_id=135,
        home_aruco_id=131,
    )
    sm.set_task(task)
    
    sm.transition_to(DeliveryState.TAKING_OFF)
    assert sm.current_task.state == DeliveryState.TAKING_OFF
    
    sm.transition_to(DeliveryState.ARRIVED)
    assert sm.current_task.state == DeliveryState.ARRIVED


def test_transition_without_task():
    sm = StateMachine()
    sm.transition_to(DeliveryState.ARRIVED)
    assert sm.current_task is None


def test_event_idempotency():
    sm = StateMachine()
    event_id = "test-event-123"
    
    assert not sm.is_event_sent(event_id)
    
    sm.mark_event_sent(event_id)
    assert sm.is_event_sent(event_id)
    
    sm.mark_event_sent(event_id)
    assert sm.is_event_sent(event_id)


def test_clear_task():
    sm = StateMachine()
    task = DeliveryTask(
        delivery_id="test-123",
        order_id="order-456",
        good_id="good-789",
        parcel_automat_id="automat-1",
        target_aruco_id=135,
        home_aruco_id=131,
    )
    sm.set_task(task)
    sm.mark_event_sent("event-1")
    sm.mark_event_sent("event-2")
    
    sm.clear_task()
    
    assert sm.current_task is None
    assert len(sm._sent_events) == 0


def test_get_state():
    sm = StateMachine()
    
    assert sm.get_state() is None
    
    task = DeliveryTask(
        delivery_id="test-123",
        order_id="order-456",
        good_id="good-789",
        parcel_automat_id="automat-1",
        target_aruco_id=135,
        home_aruco_id=131,
    )
    sm.set_task(task)
    
    assert sm.get_state() == DeliveryState.PENDING
    
    sm.transition_to(DeliveryState.NAVIGATING)
    assert sm.get_state() == DeliveryState.NAVIGATING


def test_multiple_tasks_clear_events():
    sm = StateMachine()
    
    task1 = DeliveryTask(
        delivery_id="test-123",
        order_id="order-456",
        good_id="good-789",
        parcel_automat_id="automat-1",
        target_aruco_id=135,
        home_aruco_id=131,
    )
    sm.set_task(task1)
    sm.mark_event_sent("task1-arrived")
    
    sm.clear_task()
    
    task2 = DeliveryTask(
        delivery_id="test-456",
        order_id="order-789",
        good_id="good-111",
        parcel_automat_id="automat-2",
        target_aruco_id=136,
        home_aruco_id=131,
    )
    sm.set_task(task2)
    
    assert not sm.is_event_sent("task1-arrived")
    assert sm.current_task.delivery_id == "test-456"
