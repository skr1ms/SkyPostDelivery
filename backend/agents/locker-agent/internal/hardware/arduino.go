package hardware

import (
	"fmt"
	"time"

	entityError "github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
	"go.bug.st/serial"
)

type ArduinoController struct {
	port     serial.Port
	mockMode bool
	logger   logger.Interface
}

func NewArduinoController(portName string, baudrate int, mockMode bool, log logger.Interface) (*ArduinoController, error) {
	if mockMode {
		log.Info("Arduino controller running in MOCK mode", nil)
		return &ArduinoController{
			mockMode: true,
			logger:   log,
		}, nil
	}

	mode := &serial.Mode{
		BaudRate: baudrate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Warn("Failed to open Arduino serial port, switching to MOCK mode", err, map[string]any{
			"port": portName,
		})
		return &ArduinoController{
			mockMode: true,
			logger:   log,
		}, nil
	}

	if err := port.SetReadTimeout(time.Second); err != nil {
		log.Warn("Failed to set read timeout", err)
	}

	time.Sleep(500 * time.Millisecond)

	log.Info("Arduino controller initialized", nil, map[string]any{
		"port":     portName,
		"baudrate": baudrate,
	})

	return &ArduinoController{
		port:     port,
		mockMode: false,
		logger:   log,
	}, nil
}

func (a *ArduinoController) sendCommand(command string) (string, error) {
	if a.mockMode {
		a.logger.Debug("MOCK Arduino command", nil, map[string]any{
			"command": command,
		})
		return "OK", nil
	}

	cmd := fmt.Sprintf("%s\n", command)
	_, err := a.port.Write([]byte(cmd))
	if err != nil {
		return "", entityError.ErrArduinoConnectionFailed
	}

	time.Sleep(50 * time.Millisecond)

	buf := make([]byte, 128)
	n, err := a.port.Read(buf)
	if err != nil {
		return "", entityError.ErrArduinoConnectionFailed
	}

	response := string(buf[:n])
	a.logger.Debug("Arduino response", nil, map[string]any{
		"command":  command,
		"response": response,
	})

	return response, nil
}

func (a *ArduinoController) OpenCell(cellNumber int) error {
	command := fmt.Sprintf("open_%d", cellNumber)
	response, err := a.sendCommand(command)
	if err != nil {
		return err
	}

	a.logger.Info("Cell opened", nil, map[string]any{
		"cell_number": cellNumber,
		"response":    response,
	})

	return nil
}

func (a *ArduinoController) OpenInternalDoor(doorNumber int) error {
	command := fmt.Sprintf("internal_%d", doorNumber)
	response, err := a.sendCommand(command)
	if err != nil {
		return err
	}

	a.logger.Info("Internal door opened", nil, map[string]any{
		"door_number": doorNumber,
		"response":    response,
	})

	return nil
}

func (a *ArduinoController) GetCellsCount() (int, error) {
	if a.mockMode {
		return 12, nil
	}

	response, err := a.sendCommand("cells")
	if err != nil {
		return 0, err
	}

	var count int
	if _, err := fmt.Sscanf(response, "%d", &count); err != nil {
		a.logger.Warn("Failed to parse cell count from Arduino", err, map[string]any{
			"response": response,
		})
		return 0, entityError.ErrArduinoCommandFailed
	}

	return count, nil
}

func (a *ArduinoController) Reset() error {
	_, err := a.sendCommand("reset")
	if err != nil {
		return err
	}

	a.logger.Info("Arduino reset completed", nil)
	return nil
}

func (a *ArduinoController) Close() error {
	if a.mockMode || a.port == nil {
		return nil
	}

	err := a.port.Close()
	if err != nil {
		return entityError.ErrArduinoConnectionFailed
	}

	a.logger.Info("Arduino serial port closed", nil)
	return nil
}

func (a *ArduinoController) IsMockMode() bool {
	return a.mockMode
}
