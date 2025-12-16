package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     logger.Interface
	retryCount int
}

func NewClient(baseURL string, timeout time.Duration, retryCount int, log logger.Interface) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger:     log,
		retryCount: retryCount,
	}
}

func (c *Client) ValidateQR(ctx context.Context, qrData, parcelAutomatID string) (*entity.QRValidationResponse, error) {
	url := fmt.Sprintf("%s/automats/qr-scan", c.baseURL)

	req := entity.QRValidationRequest{
		QRData:          qrData,
		ParcelAutomatID: parcelAutomatID,
	}

	var response entity.QRValidationResponse
	if err := c.doRequest(ctx, "POST", url, req, &response); err != nil {
		return nil, fmt.Errorf("OrchestratorClient - ValidateQR - doRequest: %w", err)
	}

	return &response, nil
}

func (c *Client) ConfirmPickup(ctx context.Context, cellIDs []uuid.UUID) error {
	url := fmt.Sprintf("%s/automats/confirm-pickup", c.baseURL)

	cellIDStrings := make([]string, len(cellIDs))
	for i, id := range cellIDs {
		cellIDStrings[i] = id.String()
	}

	req := entity.ConfirmPickupRequest{
		CellIDs: cellIDStrings,
	}

	if err := c.doRequest(ctx, "POST", url, req, nil); err != nil {
		return fmt.Errorf("OrchestratorClient - ConfirmPickup - doRequest: %w", err)
	}

	return nil
}

func (c *Client) ConfirmLoaded(ctx context.Context, orderID, lockerCellID uuid.UUID) error {
	url := fmt.Sprintf("%s/deliveries/confirm-loaded", c.baseURL)

	req := entity.ConfirmLoadedRequest{
		OrderID:      orderID.String(),
		LockerCellID: lockerCellID.String(),
		Timestamp:    time.Now(),
	}

	if err := c.doRequest(ctx, "POST", url, req, nil); err != nil {
		return fmt.Errorf("OrchestratorClient - ConfirmLoaded - doRequest: %w", err)
	}

	return nil
}

func (c *Client) doRequest(ctx context.Context, method, url string, reqBody, respBody any) error {
	var lastErr error

	for attempt := 0; attempt <= c.retryCount; attempt++ {
		if ctx.Err() != nil {
			return fmt.Errorf("OrchestratorClient - doRequest - context: %w", ctx.Err())
		}

		if attempt > 0 {
			c.logger.Warn("Retrying request to orchestrator", nil, map[string]any{
				"attempt": attempt,
				"url":     url,
			})

			select {
			case <-time.After(time.Second * time.Duration(attempt)):
			case <-ctx.Done():
				return fmt.Errorf("OrchestratorClient - doRequest - context: %w", ctx.Err())
			}
		}

		var bodyReader io.Reader
		if reqBody != nil {
			jsonData, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("OrchestratorClient - doRequest - Marshal: %w", err)
			}
			bodyReader = bytes.NewReader(jsonData)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			lastErr = fmt.Errorf("OrchestratorClient - doRequest - NewRequest: %w", err)
			continue
		}

		if reqBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("OrchestratorClient - doRequest - Do: %w", ctx.Err())
			}
			lastErr = fmt.Errorf("OrchestratorClient - doRequest - Do: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("OrchestratorClient - doRequest - ReadAll: %w", err)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("OrchestratorClient - doRequest - bad status: %d, body: %s", resp.StatusCode, string(body))
			continue
		}

		if respBody != nil {
			if err := json.Unmarshal(body, respBody); err != nil {
				lastErr = fmt.Errorf("OrchestratorClient - doRequest - Unmarshal: %w", err)
				continue
			}
		}

		c.logger.Debug("Orchestrator request successful", nil, map[string]any{
			"method": method,
			"url":    url,
		})

		return nil
	}

	return lastErr
}
