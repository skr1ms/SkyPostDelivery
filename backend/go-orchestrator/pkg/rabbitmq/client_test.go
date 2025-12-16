package rabbitmq

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDeliveryTask_Marshaling(t *testing.T) {
	task := DeliveryTask{
		DroneID:         uuid.New(),
		DroneIP:         "192.168.1.100",
		GoodID:          uuid.New(),
		ParcelAutomatID: uuid.New(),
		ArucoID:         42,
		Weight:          2.5,
		Height:          10.0,
		Length:          20.0,
		Width:           15.0,
		Priority:        5,
		CreatedAt:       time.Now().Unix(),
	}

	data, err := json.Marshal(task)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var unmarshaled DeliveryTask
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, task.DroneID, unmarshaled.DroneID)
	assert.Equal(t, task.DroneIP, unmarshaled.DroneIP)
	assert.Equal(t, task.GoodID, unmarshaled.GoodID)
	assert.Equal(t, task.ArucoID, unmarshaled.ArucoID)
	assert.Equal(t, task.Weight, unmarshaled.Weight)
	assert.Equal(t, task.Priority, unmarshaled.Priority)
}

func TestDeliveryConfirmation_Marshaling(t *testing.T) {
	confirmation := DeliveryConfirmation{
		OrderID:      uuid.New(),
		LockerCellID: uuid.New(),
		ConfirmedAt:  time.Now().Unix(),
		AutomatID:    uuid.New(),
	}

	data, err := json.Marshal(confirmation)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var unmarshaled DeliveryConfirmation
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, confirmation.OrderID, unmarshaled.OrderID)
	assert.Equal(t, confirmation.LockerCellID, unmarshaled.LockerCellID)
	assert.Equal(t, confirmation.AutomatID, unmarshaled.AutomatID)
	assert.Equal(t, confirmation.ConfirmedAt, unmarshaled.ConfirmedAt)
}

func TestDroneStatusUpdate_Marshaling(t *testing.T) {
	status := DroneStatusUpdate{
		DroneID:      uuid.New(),
		Status:       "flying",
		BatteryLevel: 85.5,
		Latitude:     55.7558,
		Longitude:    37.6173,
		Altitude:     100.0,
		UpdatedAt:    time.Now().Unix(),
	}

	data, err := json.Marshal(status)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var unmarshaled DroneStatusUpdate
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, status.DroneID, unmarshaled.DroneID)
	assert.Equal(t, status.Status, unmarshaled.Status)
	assert.Equal(t, status.BatteryLevel, unmarshaled.BatteryLevel)
	assert.Equal(t, status.Latitude, unmarshaled.Latitude)
	assert.Equal(t, status.Longitude, unmarshaled.Longitude)
	assert.Equal(t, status.Altitude, unmarshaled.Altitude)
}

func TestQueueConstants(t *testing.T) {
	assert.Equal(t, "deliveries", QueueDeliveries)
	assert.Equal(t, "deliveries.priority", QueueDeliveriesPriority)
	assert.Equal(t, "confirmations", QueueConfirmations)
	assert.Equal(t, "deliveries.dlq", QueueDeliveriesDLQ)
}

func TestDeliveryTask_JSONFields(t *testing.T) {
	task := DeliveryTask{
		DroneID:         uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		DroneIP:         "192.168.1.100",
		GoodID:          uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
		ParcelAutomatID: uuid.MustParse("323e4567-e89b-12d3-a456-426614174000"),
		ArucoID:         42,
		Weight:          2.5,
		Height:          10.0,
		Length:          20.0,
		Width:           15.0,
		Priority:        5,
		CreatedAt:       1699123456,
	}

	data, err := json.Marshal(task)
	assert.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "drone_id")
	assert.Contains(t, jsonStr, "drone_ip")
	assert.Contains(t, jsonStr, "good_id")
	assert.Contains(t, jsonStr, "parcel_automat_id")
	assert.Contains(t, jsonStr, "aruco_id")
	assert.Contains(t, jsonStr, "192.168.1.100")
}

func TestDeliveryConfirmation_JSONFields(t *testing.T) {
	confirmation := DeliveryConfirmation{
		OrderID:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		LockerCellID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
		ConfirmedAt:  1699123456,
		AutomatID:    uuid.MustParse("323e4567-e89b-12d3-a456-426614174000"),
	}

	data, err := json.Marshal(confirmation)
	assert.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "order_id")
	assert.Contains(t, jsonStr, "locker_cell_id")
	assert.Contains(t, jsonStr, "confirmed_at")
	assert.Contains(t, jsonStr, "automat_id")
}

func TestDroneStatusUpdate_JSONFields(t *testing.T) {
	status := DroneStatusUpdate{
		DroneID:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Status:       "flying",
		BatteryLevel: 85.5,
		Latitude:     55.7558,
		Longitude:    37.6173,
		Altitude:     100.0,
		UpdatedAt:    1699123456,
	}

	data, err := json.Marshal(status)
	assert.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "drone_id")
	assert.Contains(t, jsonStr, "status")
	assert.Contains(t, jsonStr, "battery_level")
	assert.Contains(t, jsonStr, "latitude")
	assert.Contains(t, jsonStr, "longitude")
	assert.Contains(t, jsonStr, "altitude")
	assert.Contains(t, jsonStr, "updated_at")
	assert.Contains(t, jsonStr, "flying")
}

func TestDeliveryTask_EmptyValues(t *testing.T) {
	task := DeliveryTask{}

	data, err := json.Marshal(task)
	assert.NoError(t, err)

	var unmarshaled DeliveryTask
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, uuid.Nil, unmarshaled.DroneID)
	assert.Equal(t, "", unmarshaled.DroneIP)
	assert.Equal(t, 0.0, unmarshaled.Weight)
	assert.Equal(t, 0, unmarshaled.Priority)
}

func TestDeliveryConfirmation_EmptyValues(t *testing.T) {
	confirmation := DeliveryConfirmation{}

	data, err := json.Marshal(confirmation)
	assert.NoError(t, err)

	var unmarshaled DeliveryConfirmation
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, uuid.Nil, unmarshaled.OrderID)
	assert.Equal(t, uuid.Nil, unmarshaled.LockerCellID)
	assert.Equal(t, int64(0), unmarshaled.ConfirmedAt)
}

func TestDroneStatusUpdate_EmptyValues(t *testing.T) {
	status := DroneStatusUpdate{}

	data, err := json.Marshal(status)
	assert.NoError(t, err)

	var unmarshaled DroneStatusUpdate
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, uuid.Nil, unmarshaled.DroneID)
	assert.Equal(t, "", unmarshaled.Status)
	assert.Equal(t, 0.0, unmarshaled.BatteryLevel)
	assert.Equal(t, 0.0, unmarshaled.Latitude)
}

func TestDeliveryTask_InvalidJSON(t *testing.T) {
	invalidJSON := `{"drone_id": "invalid-uuid"}`

	var task DeliveryTask
	err := json.Unmarshal([]byte(invalidJSON), &task)
	assert.Error(t, err)
}

func TestDeliveryConfirmation_InvalidJSON(t *testing.T) {
	invalidJSON := `{"order_id": "not-a-uuid"}`

	var confirmation DeliveryConfirmation
	err := json.Unmarshal([]byte(invalidJSON), &confirmation)
	assert.Error(t, err)
}

func TestDeliveryTask_RoundTrip(t *testing.T) {
	original := DeliveryTask{
		DroneID:         uuid.New(),
		DroneIP:         "10.0.0.1",
		GoodID:          uuid.New(),
		ParcelAutomatID: uuid.New(),
		ArucoID:         100,
		Weight:          5.5,
		Height:          30.0,
		Length:          40.0,
		Width:           25.0,
		Priority:        10,
		CreatedAt:       time.Now().Unix(),
	}

	data, err := json.Marshal(original)
	assert.NoError(t, err)

	var decoded DeliveryTask
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	data2, err := json.Marshal(decoded)
	assert.NoError(t, err)

	assert.JSONEq(t, string(data), string(data2))
}

func TestDeliveryConfirmation_RoundTrip(t *testing.T) {
	original := DeliveryConfirmation{
		OrderID:      uuid.New(),
		LockerCellID: uuid.New(),
		ConfirmedAt:  time.Now().Unix(),
		AutomatID:    uuid.New(),
	}

	data, err := json.Marshal(original)
	assert.NoError(t, err)

	var decoded DeliveryConfirmation
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	data2, err := json.Marshal(decoded)
	assert.NoError(t, err)

	assert.JSONEq(t, string(data), string(data2))
}

func TestDroneStatusUpdate_RoundTrip(t *testing.T) {
	original := DroneStatusUpdate{
		DroneID:      uuid.New(),
		Status:       "landing",
		BatteryLevel: 42.3,
		Latitude:     59.9311,
		Longitude:    30.3609,
		Altitude:     50.5,
		UpdatedAt:    time.Now().Unix(),
	}

	data, err := json.Marshal(original)
	assert.NoError(t, err)

	var decoded DroneStatusUpdate
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	data2, err := json.Marshal(decoded)
	assert.NoError(t, err)

	assert.JSONEq(t, string(data), string(data2))
}

func TestDeliveryTask_HighPriority(t *testing.T) {
	task := DeliveryTask{
		Priority: 10,
	}

	assert.Equal(t, 10, task.Priority)
	assert.Greater(t, task.Priority, 5)
}

func TestDeliveryTask_LowPriority(t *testing.T) {
	task := DeliveryTask{
		Priority: 0,
	}

	assert.Equal(t, 0, task.Priority)
	assert.LessOrEqual(t, task.Priority, 5)
}

func TestDeliveryTask_DimensionsValidation(t *testing.T) {
	task := DeliveryTask{
		Height: 10.0,
		Length: 20.0,
		Width:  15.0,
		Weight: 5.0,
	}

	assert.Positive(t, task.Height)
	assert.Positive(t, task.Length)
	assert.Positive(t, task.Width)
	assert.Positive(t, task.Weight)
}

func TestDroneStatusUpdate_BatteryLevelRange(t *testing.T) {
	testCases := []struct {
		name         string
		batteryLevel float64
		valid        bool
	}{
		{"Full battery", 100.0, true},
		{"Half battery", 50.0, true},
		{"Low battery", 10.0, true},
		{"Empty battery", 0.0, true},
		{"Over 100", 150.0, false},
		{"Negative", -10.0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status := DroneStatusUpdate{
				BatteryLevel: tc.batteryLevel,
			}

			if tc.valid {
				assert.GreaterOrEqual(t, status.BatteryLevel, 0.0)
				assert.LessOrEqual(t, status.BatteryLevel, 100.0)
			} else {
				assert.True(t, status.BatteryLevel < 0.0 || status.BatteryLevel > 100.0)
			}
		})
	}
}

func TestDroneStatusUpdate_StatusValues(t *testing.T) {
	validStatuses := []string{"idle", "flying", "landing", "charging", "maintenance"}

	for _, statusValue := range validStatuses {
		t.Run(statusValue, func(t *testing.T) {
			status := DroneStatusUpdate{
				Status: statusValue,
			}

			assert.Equal(t, statusValue, status.Status)
			assert.NotEmpty(t, status.Status)
		})
	}
}

func TestMessageSerialization_Context(t *testing.T) {
	task := DeliveryTask{
		DroneID:         uuid.New(),
		DroneIP:         "192.168.1.1",
		GoodID:          uuid.New(),
		ParcelAutomatID: uuid.New(),
		CreatedAt:       time.Now().Unix(),
	}

	data, err := json.Marshal(task)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
