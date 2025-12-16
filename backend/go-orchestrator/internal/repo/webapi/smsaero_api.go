package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	webapierror "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/webapi/error"
)

type SMSAeroAPI struct {
	email   string
	apiKey  string
	baseURL string
	client  *http.Client
}

type SMSAeroResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func NewSMSAeroAPI(email, apiKey, baseURL string) *SMSAeroAPI {
	return &SMSAeroAPI{
		email:   email,
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (s *SMSAeroAPI) SendSMS(ctx context.Context, phone, code string) error {
	endpoint := fmt.Sprintf("%s/sms/send", s.baseURL)

	data := url.Values{}
	data.Set("number", phone)
	data.Set("text", fmt.Sprintf("Ваш код для входа в SkyPost Delivery: %s", code))
	data.Set("sign", "SMS Aero")

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("SMSAeroAPI - SendSMS - NewRequest: %w", err)
	}

	req.SetBasicAuth(s.email, s.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("SMSAeroAPI - SendSMS - Do: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("SMSAeroAPI - SendSMS - ReadAll: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusBadRequest:
		return fmt.Errorf("SMSAeroAPI - SendSMS - HandleResponse[status=%d]: %w", resp.StatusCode, webapierror.ErrSMSInvalidPhone)
	case http.StatusPaymentRequired:
		return webapierror.ErrSMSInsufficientFunds
	case http.StatusTooManyRequests:
		return webapierror.ErrSMSRateLimitExceeded
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("SMSAeroAPI - SendSMS - HandleResponse[status=%d]: %w", resp.StatusCode, webapierror.ErrSMSServiceUnavailable)
	default:
		return fmt.Errorf("SMSAeroAPI - SendSMS - HandleResponse[status=%d]: %w", resp.StatusCode, webapierror.ErrSMSSendFailed)
	}

	var smsResp SMSAeroResponse
	if err := json.Unmarshal(body, &smsResp); err != nil {
		return fmt.Errorf("SMSAeroAPI - SendSMS - Unmarshal: %w", err)
	}

	if !smsResp.Success {
		message := strings.ToLower(smsResp.Message)
		switch {
		case strings.Contains(message, "balance") || strings.Contains(message, "funds"):
			return fmt.Errorf("SMSAeroAPI - SendSMS - HandleResponse[message=%s]: %w", smsResp.Message, webapierror.ErrSMSInsufficientFunds)
		case strings.Contains(message, "phone") || strings.Contains(message, "number"):
			return fmt.Errorf("SMSAeroAPI - SendSMS - HandleResponse[message=%s]: %w", smsResp.Message, webapierror.ErrSMSInvalidPhone)
		default:
			return fmt.Errorf("SMSAeroAPI - SendSMS - HandleResponse[message=%s]: %w", smsResp.Message, webapierror.ErrSMSSendFailed)
		}
	}

	return nil
}

func (s *SMSAeroAPI) CheckBalance(ctx context.Context) (float64, error) {
	endpoint := fmt.Sprintf("%s/balance", s.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("SMSAeroAPI - CheckBalance - NewRequest: %w", err)
	}

	req.SetBasicAuth(s.email, s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("SMSAeroAPI - CheckBalance - Do: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("SMSAeroAPI - CheckBalance - ReadAll: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("SMSAeroAPI - CheckBalance - HandleResponse[status=%d]: %w", resp.StatusCode, webapierror.ErrSMSServiceUnavailable)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Balance float64 `json:"balance"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("SMSAeroAPI - CheckBalance - Unmarshal: %w", err)
	}

	if !result.Success {
		return 0, webapierror.ErrSMSServiceUnavailable
	}

	return result.Data.Balance, nil
}
