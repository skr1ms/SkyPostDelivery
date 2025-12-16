package hardware

import (
	"fmt"
	"time"

	entityError "github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
	"go.bug.st/serial"
)

type DisplayController struct {
	port     serial.Port
	mockMode bool
	logger   logger.Interface
}

func NewDisplayController(portName string, baudrate int, mockMode bool, log logger.Interface) (*DisplayController, error) {
	if mockMode {
		log.Info("Display controller running in MOCK mode", nil)
		return &DisplayController{
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
		log.Warn("Failed to open display serial port, switching to MOCK mode", err, map[string]any{
			"port": portName,
		})
		return &DisplayController{
			mockMode: true,
			logger:   log,
		}, nil
	}

	if err := port.SetReadTimeout(time.Second); err != nil {
		log.Warn("Failed to set read timeout", err)
	}

	time.Sleep(500 * time.Millisecond)

	log.Info("Display controller initialized", nil, map[string]any{
		"port":     portName,
		"baudrate": baudrate,
	})

	return &DisplayController{
		port:     port,
		mockMode: false,
		logger:   log,
	}, nil
}

func (d *DisplayController) SendMessage(message string) error {
	if d.mockMode {
		d.logger.Info("MOCK Display message", nil, map[string]any{
			"message": message,
		})
		return nil
	}

	msg := fmt.Sprintf("%s\n", message)
	_, err := d.port.Write([]byte(msg))
	if err != nil {
		return entityError.ErrDisplayConnectionFailed
	}

	d.logger.Debug("Display message sent", nil, map[string]any{
		"message": message,
	})

	return nil
}

func (d *DisplayController) ShowWelcome() error {
	return d.SendMessage("Welcome!\nScan your QR")
}

func (d *DisplayController) ShowScanning() error {
	return d.SendMessage("Scanning...\nPlease wait")
}

func (d *DisplayController) ShowSuccess(customerName string) error {
	if customerName != "" {
		return d.SendMessage(fmt.Sprintf("Hello,\n%s!", customerName))
	}
	return d.SendMessage("QR accepted!\nOpening cell...")
}

func (d *DisplayController) ShowInvalid() error {
	return d.SendMessage("Invalid QR!\nTry again")
}

func (d *DisplayController) ShowCellOpening(cellNumber int, orderNumber string) error {
	if orderNumber != "" {
		return d.SendMessage(fmt.Sprintf("Order #%s\nCell #%d", orderNumber, cellNumber))
	}
	return d.SendMessage(fmt.Sprintf("Your cell:\n#%d", cellNumber))
}

func (d *DisplayController) ShowCellOpened(cellNumber int) error {
	return d.SendMessage(fmt.Sprintf("Cell #%d\nOPEN", cellNumber))
}

func (d *DisplayController) ShowPleaseClose() error {
	return d.SendMessage("Take your item\nClose the door")
}

func (d *DisplayController) ShowThankYou() error {
	return d.SendMessage("Thank you!\nHave a nice day")
}

func (d *DisplayController) ShowError(message string) error {
	if message == "" {
		message = "Error occurred"
	}
	return d.SendMessage(fmt.Sprintf("Error:\n%s", message))
}

func (d *DisplayController) Close() error {
	if d.mockMode || d.port == nil {
		return nil
	}

	err := d.port.Close()
	if err != nil {
		return entityError.ErrDisplayConnectionFailed
	}

	d.logger.Info("Display serial port closed", nil)
	return nil
}

func (d *DisplayController) IsMockMode() bool {
	return d.mockMode
}
