package hardware

import (
	"context"
	"fmt"
	"time"

	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
	"gocv.io/x/gocv"
)

type QRCamera struct {
	camera       *gocv.VideoCapture
	detector     *gocv.QRCodeDetector
	mockMode     bool
	logger       logger.Interface
	scanInterval time.Duration
	resultChan   chan string
	stopChan     chan struct{}
	running      bool
}

func NewQRCamera(cameraIndex, width, height, fps, scanIntervalMS int, mockMode bool, log logger.Interface) (*QRCamera, error) {
	if mockMode {
		log.Info("QR camera running in MOCK mode", nil)
		return &QRCamera{
			mockMode:     true,
			logger:       log,
			scanInterval: time.Duration(scanIntervalMS) * time.Millisecond,
			resultChan:   make(chan string, 10),
			stopChan:     make(chan struct{}),
		}, nil
	}

	camera, err := gocv.OpenVideoCapture(cameraIndex)
	if err != nil {
		log.Warn("Failed to open camera, switching to MOCK mode", err, map[string]interface{}{
			"camera_index": cameraIndex,
		})
		return &QRCamera{
			mockMode:     true,
			logger:       log,
			scanInterval: time.Duration(scanIntervalMS) * time.Millisecond,
			resultChan:   make(chan string, 10),
			stopChan:     make(chan struct{}),
		}, nil
	}

	camera.Set(gocv.VideoCaptureFrameWidth, float64(width))
	camera.Set(gocv.VideoCaptureFrameHeight, float64(height))
	camera.Set(gocv.VideoCaptureFPS, float64(fps))

	if !camera.IsOpened() {
		log.Warn("Camera opened but not ready, switching to MOCK mode", nil)
		if err := camera.Close(); err != nil {
			log.Error("Failed to close camera", err)
		}
		return &QRCamera{
			mockMode:     true,
			logger:       log,
			scanInterval: time.Duration(scanIntervalMS) * time.Millisecond,
			resultChan:   make(chan string, 10),
			stopChan:     make(chan struct{}),
		}, nil
	}

	detector := gocv.NewQRCodeDetector()

	log.Info("QR camera initialized", nil, map[string]interface{}{
		"camera_index": cameraIndex,
		"width":        width,
		"height":       height,
		"fps":          fps,
	})

	return &QRCamera{
		camera:       camera,
		detector:     &detector,
		mockMode:     false,
		logger:       log,
		scanInterval: time.Duration(scanIntervalMS) * time.Millisecond,
		resultChan:   make(chan string, 10),
		stopChan:     make(chan struct{}),
	}, nil
}

func (q *QRCamera) Start(ctx context.Context) {
	if q.running {
		return
	}

	q.running = true
	q.logger.Info("QR camera scanner started", nil, map[string]interface{}{
		"mock_mode": q.mockMode,
		"interval":  q.scanInterval.String(),
	})

	go q.scanLoop(ctx)
}

func (q *QRCamera) scanLoop(ctx context.Context) {
	ticker := time.NewTicker(q.scanInterval)
	defer ticker.Stop()

	img := gocv.NewMat()
	defer func() {
		if err := img.Close(); err != nil {
			q.logger.Error("Failed to close image mat", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			q.logger.Info("QR camera scanner stopped by context", nil)
			return
		case <-q.stopChan:
			q.logger.Info("QR camera scanner stopped", nil)
			return
		case <-ticker.C:
			if q.mockMode {
				continue
			}

			if ok := q.camera.Read(&img); !ok || img.Empty() {
				continue
			}

			points := gocv.NewMat()
			straight_qrcode := gocv.NewMat()
			data := q.detector.DetectAndDecode(img, &points, &straight_qrcode)
			if err := points.Close(); err != nil {
				q.logger.Error("Failed to close points", err)
			}
			if err := straight_qrcode.Close(); err != nil {
				q.logger.Error("Failed to close straight_qrcode", err)
			}
			if data != "" {
				q.logger.Debug("QR code detected", nil, map[string]interface{}{
					"data_length": len(data),
				})

				select {
				case q.resultChan <- data:
				default:
					q.logger.Warn("QR result channel full, dropping scan result", nil)
				}
			}
		}
	}
}

func (q *QRCamera) Stop() {
	if !q.running {
		return
	}

	close(q.stopChan)
	q.running = false
	q.logger.Info("QR camera scanner stopping", nil)
}

func (q *QRCamera) GetResultChannel() <-chan string {
	return q.resultChan
}

func (q *QRCamera) ScanOnce() (string, error) {
	if q.mockMode {
		return "", fmt.Errorf("qr camera - ScanOnce: mock mode enabled")
	}

	img := gocv.NewMat()
	defer func() {
		if err := img.Close(); err != nil {
			q.logger.Error("Failed to close image mat", err)
		}
	}()

	if ok := q.camera.Read(&img); !ok || img.Empty() {
		return "", fmt.Errorf("qr camera - ScanOnce: failed to read frame")
	}

	points := gocv.NewMat()
	defer func() {
		if err := points.Close(); err != nil {
			q.logger.Error("Failed to close points", err)
		}
	}()
	straight_qrcode := gocv.NewMat()
	defer func() {
		if err := straight_qrcode.Close(); err != nil {
			q.logger.Error("Failed to close straight_qrcode", err)
		}
	}()
	data := q.detector.DetectAndDecode(img, &points, &straight_qrcode)
	if data == "" {
		return "", fmt.Errorf("qr camera - ScanOnce: no QR code detected")
	}

	return data, nil
}

func (q *QRCamera) InjectMockQR(qrData string) error {
	if !q.mockMode {
		return fmt.Errorf("qr camera - InjectMockQR: not in mock mode")
	}

	select {
	case q.resultChan <- qrData:
		q.logger.Info("Mock QR injected", nil, map[string]interface{}{
			"qr_data": qrData,
		})
		return nil
	default:
		return fmt.Errorf("qr camera - InjectMockQR: result channel full")
	}
}

func (q *QRCamera) Close() error {
	q.Stop()

	if q.mockMode {
		return nil
	}

	if q.camera != nil {
		if err := q.camera.Close(); err != nil {
			return fmt.Errorf("qr camera - Close: %w", err)
		}
	}

	if q.detector != nil {
		if err := q.detector.Close(); err != nil {
			q.logger.Error("Failed to close detector", err)
		}
	}

	q.logger.Info("QR camera closed", nil)
	return nil
}

func (q *QRCamera) IsMockMode() bool {
	return q.mockMode
}

func (q *QRCamera) IsRunning() bool {
	return q.running
}
