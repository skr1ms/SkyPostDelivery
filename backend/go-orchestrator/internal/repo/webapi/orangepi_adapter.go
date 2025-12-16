package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	webapierror "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/webapi/error"
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
	CellsOut        []string `json:"cells_out"`
	CellsInternal   []string `json:"cells_internal"`
	ParcelAutomatID string   `json:"parcel_automat_id"`
}

func (a *OrangePIAdapter) SendCellUUIDs(ctx context.Context, ipAddress string, parcelAutomatID uuid.UUID, outCellUUIDs []uuid.UUID, internalCellUUIDs []uuid.UUID) error {
	if ipAddress == "" {
		return webapierror.ErrOrangePIEmptyIP
	}

	outStrings := make([]string, len(outCellUUIDs))
	for i, id := range outCellUUIDs {
		outStrings[i] = id.String()
	}

	internalStrings := make([]string, len(internalCellUUIDs))
	for i, id := range internalCellUUIDs {
		internalStrings[i] = id.String()
	}

	payload := CellUUIDsPayload{
		CellsOut:        outStrings,
		CellsInternal:   internalStrings,
		ParcelAutomatID: parcelAutomatID.String(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("OrangePIAdapter - SendCellUUIDs - Marshal: %w", err)
	}

	url := fmt.Sprintf("http://%s:8000/api/cells/sync", ipAddress)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("OrangePIAdapter - SendCellUUIDs - NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("OrangePIAdapter - SendCellUUIDs - Do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("OrangePIAdapter - SendCellUUIDs - ReadAll: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest, http.StatusNotFound:
		return fmt.Errorf("OrangePIAdapter - SendCellUUIDs - HandleResponse[status=%d, body=%s]: %w", resp.StatusCode, string(body), webapierror.ErrOrangePISendFailed)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("OrangePIAdapter - SendCellUUIDs - HandleResponse[status=%d]: %w", resp.StatusCode, webapierror.ErrOrangePIServiceUnavailable)
	default:
		return fmt.Errorf("OrangePIAdapter - SendCellUUIDs - HandleResponse[status=%d, body=%s]: %w", resp.StatusCode, string(body), webapierror.ErrOrangePISendFailed)
	}
}

type OpenCellPayload struct {
	CellID string `json:"cell_id"`
}

func (a *OrangePIAdapter) OpenCell(ctx context.Context, ipAddress string, cellID uuid.UUID) error {
	if ipAddress == "" {
		return webapierror.ErrOrangePIEmptyIP
	}

	payload := OpenCellPayload{
		CellID: cellID.String(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("OrangePIAdapter - OpenCell - Marshal: %w", err)
	}

	url := fmt.Sprintf("http://%s:8000/api/cells/prepare", ipAddress)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("OrangePIAdapter - OpenCell - NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("OrangePIAdapter - OpenCell - Do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("OrangePIAdapter - OpenCell - ReadAll: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest, http.StatusNotFound:
		return fmt.Errorf("OrangePIAdapter - OpenCell - HandleResponse[status=%d, body=%s]: %w", resp.StatusCode, string(body), webapierror.ErrOrangePIOpenCellFailed)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("OrangePIAdapter - OpenCell - HandleResponse[status=%d]: %w", resp.StatusCode, webapierror.ErrOrangePIServiceUnavailable)
	default:
		return fmt.Errorf("OrangePIAdapter - OpenCell - HandleResponse[status=%d, body=%s]: %w", resp.StatusCode, string(body), webapierror.ErrOrangePIOpenCellFailed)
	}
}
