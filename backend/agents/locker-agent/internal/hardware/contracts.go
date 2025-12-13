package hardware

import "context"

type ArduinoInterface interface {
	OpenCell(cellNumber int) error
	OpenInternalDoor(doorNumber int) error
	GetCellsCount() (int, error)
	Close() error
	IsMockMode() bool
}

type DisplayInterface interface {
	ShowCellOpening(cellNumber int, orderNumber string) error
	ShowCellOpened(cellNumber int) error
	ShowScanning() error
	ShowSuccess(message string) error
	ShowInvalid() error
	ShowError(message string) error
	ShowPleaseClose() error
	ShowThankYou() error
	Close() error
	IsMockMode() bool
}

type QRCameraInterface interface {
	Start(ctx context.Context)
	Stop()
	GetResultChannel() <-chan string
	ScanOnce() (string, error)
	IsMockMode() bool
}
