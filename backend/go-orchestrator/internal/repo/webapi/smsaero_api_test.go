package webapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSMSAeroAPI(t *testing.T) {
	email := "test@example.com"
	apiKey := "test-api-key"
	baseURL := "https://gate.smsaero.ru/v2"

	api := NewSMSAeroAPI(email, apiKey, baseURL)

	assert.NotNil(t, api)
	assert.Equal(t, email, api.email)
	assert.Equal(t, apiKey, api.apiKey)
	assert.Equal(t, "https://gate.smsaero.ru/v2", api.baseURL)
	assert.NotNil(t, api.client)
}

func TestSMSAeroAPI_SendSMS_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/sms/send", r.URL.Path)

		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "test@example.com", username)
		assert.Equal(t, "test-api-key", password)

		err := r.ParseForm()
		assert.NoError(t, err)
		assert.Equal(t, "79991234567", r.FormValue("number"))
		assert.Equal(t, "Ваш код для входа в SkyPost Delivery: Test message", r.FormValue("text"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true, "message": "SMS sent"}`))
	}))
	defer server.Close()

	email := "test@example.com"
	apiKey := "test-api-key"

	api := NewSMSAeroAPI(email, apiKey, server.URL)

	err := api.SendSMS(context.Background(), "79991234567", "Test message")

	assert.NoError(t, err)
}

func TestSMSAeroAPI_SendSMS_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": false, "message": "Insufficient balance"}`))
	}))
	defer server.Close()

	email := "test@example.com"
	apiKey := "test-api-key"

	api := NewSMSAeroAPI(email, apiKey, server.URL)

	err := api.SendSMS(context.Background(), "79991234567", "Test message")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Insufficient balance")
}

func TestSMSAeroAPI_SendSMS_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Unauthorized"}`))
	}))
	defer server.Close()

	email := "test@example.com"
	apiKey := "test-api-key"

	api := NewSMSAeroAPI(email, apiKey, server.URL)

	err := api.SendSMS(context.Background(), "79991234567", "Test message")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMSAeroAPI - SendSMS - HandleResponse[status=401]")
}

func TestSMSAeroAPI_SendSMS_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	email := "test@example.com"
	apiKey := "test-api-key"

	api := NewSMSAeroAPI(email, apiKey, server.URL)

	err := api.SendSMS(context.Background(), "79991234567", "Test message")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMSAeroAPI - SendSMS - Unmarshal")
}

func TestSMSAeroAPI_CheckBalance_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/balance", r.URL.Path)

		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "test@example.com", username)
		assert.Equal(t, "test-api-key", password)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true, "data": {"balance": 100.50}}`))
	}))
	defer server.Close()

	email := "test@example.com"
	apiKey := "test-api-key"

	api := NewSMSAeroAPI(email, apiKey, server.URL)

	balance, err := api.CheckBalance(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 100.50, balance)
}

func TestSMSAeroAPI_CheckBalance_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	email := "test@example.com"
	apiKey := "test-api-key"

	api := NewSMSAeroAPI(email, apiKey, server.URL)

	balance, err := api.CheckBalance(context.Background())

	assert.Error(t, err)
	assert.Equal(t, 0.0, balance)
	assert.Contains(t, err.Error(), "SMSAeroAPI - CheckBalance - HandleResponse[status=500]")
}

func TestSMSAeroAPI_CheckBalance_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid}`))
	}))
	defer server.Close()

	email := "test@example.com"
	apiKey := "test-api-key"

	api := NewSMSAeroAPI(email, apiKey, server.URL)

	balance, err := api.CheckBalance(context.Background())

	assert.Error(t, err)
	assert.Equal(t, 0.0, balance)
}
