package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("SMSAeroAPI - SendSMS - ReadAll: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SMSAeroAPI - SendSMS - status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var smsResp SMSAeroResponse
	if err := json.Unmarshal(body, &smsResp); err != nil {
		return fmt.Errorf("SMSAeroAPI - SendSMS - Unmarshal: %w", err)
	}

	if !smsResp.Success {
		return fmt.Errorf("SMSAeroAPI - SendSMS - API error: %s", smsResp.Message)
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("SMSAeroAPI - CheckBalance - ReadAll: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("SMSAeroAPI - CheckBalance - status code: %d", resp.StatusCode)
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

	return result.Data.Balance, nil
}
