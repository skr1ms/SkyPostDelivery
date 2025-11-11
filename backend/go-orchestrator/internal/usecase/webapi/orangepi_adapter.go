package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type OrangePIAdapter struct {
	httpClient *http.Client
}

func NewOrangePIAdapter() *OrangePIAdapter {
	return &OrangePIAdapter{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type CellUUIDsPayload struct {
	Cells            []string `json:"cells"`
	ParcelAutomatID string   `json:"parcel_automat_id"`
}

func (a *OrangePIAdapter) SendCellUUIDs(ctx context.Context, ipAddress string, parcelAutomatID uuid.UUID, cellUUIDs []uuid.UUID) error {
	if ipAddress == "" {
		return fmt.Errorf("ip address is empty")
	}

	cellStrings := make([]string, len(cellUUIDs))
	for i, id := range cellUUIDs {
		cellStrings[i] = id.String()
	}

	payload := CellUUIDsPayload{
		Cells:            cellStrings,
		ParcelAutomatID: parcelAutomatID.String(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("http://%s:8000/api/cells/sync", ipAddress)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to OrangePI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OrangePI returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}

type OpenCellPayload struct {
	CellID string `json:"cell_id"`
}

func (a *OrangePIAdapter) OpenCell(ctx context.Context, ipAddress string, cellID uuid.UUID) error {
	if ipAddress == "" {
		return fmt.Errorf("ip address is empty")
	}

	payload := OpenCellPayload{
		CellID: cellID.String(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("http://%s:8000/api/cells/prepare", ipAddress)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to OrangePI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OrangePI returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}
