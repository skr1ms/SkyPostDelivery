package hardware

import (
	"context"
	"sync"
	"time"

	entityError "github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity/error"
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
	mu           sync.Mutex
	stopOnce     sync.Once
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
		log.Warn("Failed to open camera, switching to MOCK mode", err, map[string]any{
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

	log.Info("QR camera initialized", nil, map[string]any{
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
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		return
	}

	q.stopChan = make(chan struct{})
	q.stopOnce = sync.Once{}
	q.running = true
	q.logger.Info("QR camera scanner started", nil, map[string]any{
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

			q.scanFrame(&img)
		}
	}
}

func (q *QRCamera) scanFrame(img *gocv.Mat) {
	if ok := q.camera.Read(img); !ok || img.Empty() {
		return
	}

	points := gocv.NewMat()
	defer func() {
		if err := points.Close(); err != nil {
			q.logger.Error("Failed to close points", err)
		}
	}()

	straightQRCode := gocv.NewMat()
	defer func() {
		if err := straightQRCode.Close(); err != nil {
			q.logger.Error("Failed to close straight_qrcode", err)
		}
	}()

	data := q.detector.DetectAndDecode(*img, &points, &straightQRCode)
	if data == "" {
		return
	}

	q.logger.Debug("QR code detected", nil, map[string]any{
		"data_length": len(data),
	})

	select {
	case q.resultChan <- data:
	default:
		q.logger.Warn("QR result channel full, dropping scan result", nil)
	}
}

func (q *QRCamera) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running {
		return
	}

	q.stopOnce.Do(func() {
		close(q.stopChan)
	})
	q.running = false
	q.logger.Info("QR camera scanner stopping", nil)
}

func (q *QRCamera) GetResultChannel() <-chan string {
	return q.resultChan
}

func (q *QRCamera) ScanOnce() (string, error) {
	if q.mockMode {
		return "", entityError.ErrCameraNotInMockMode
	}

	img := gocv.NewMat()
	defer func() {
		if err := img.Close(); err != nil {
			q.logger.Error("Failed to close image mat", err)
		}
	}()

	if ok := q.camera.Read(&img); !ok || img.Empty() {
		return "", entityError.ErrCameraScanFailed
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
		return "", entityError.ErrCameraScanFailed
	}

	return data, nil
}

func (q *QRCamera) InjectMockQR(qrData string) error {
	if !q.mockMode {
		return entityError.ErrCameraNotInMockMode
	}

	select {
	case q.resultChan <- qrData:
		q.logger.Info("Mock QR injected", nil, map[string]any{
			"qr_data": qrData,
		})
		return nil
	default:
		return entityError.ErrCameraChannelFull
	}
}

func (q *QRCamera) Close() error {
	q.Stop()

	if q.mockMode {
		return nil
	}

	if q.camera != nil {
		if err := q.camera.Close(); err != nil {
			return entityError.ErrCameraConnectionFailed
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
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.running
}
