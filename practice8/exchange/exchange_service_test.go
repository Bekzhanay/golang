package exchange

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ─── Success ──────────────────────────────────────────────────────────────────

func TestGetRate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"base":"USD","target":"EUR","rate":0.92}`))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	rate, err := svc.GetRate("USD", "EUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate != 0.92 {
		t.Errorf("expected rate 0.92, got %f", rate)
	}
}

// ─── API Business Error (404 / 400 with JSON error) ───────────────────────────

func TestGetRate_APIBusinessError_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"invalid currency pair"}`))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "XYZ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid currency pair") {
		t.Errorf("expected 'invalid currency pair' in error, got: %v", err)
	}
}

func TestGetRate_APIBusinessError_400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid currency pair"}`))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid currency pair") {
		t.Errorf("expected 'invalid currency pair' in error, got: %v", err)
	}
}

// ─── Malformed JSON ───────────────────────────────────────────────────────────

func TestGetRate_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`Internal Server Error`)) // not valid JSON
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	if !strings.Contains(err.Error(), "decode error") {
		t.Errorf("expected 'decode error' in error, got: %v", err)
	}
}

// ─── Slow Response / Timeout ──────────────────────────────────────────────────

func TestGetRate_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than client timeout
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte(`{"rate":1.0}`))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	svc.Client = &http.Client{Timeout: 50 * time.Millisecond} // very short timeout

	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("expected 'network error' in error, got: %v", err)
	}
}

// ─── Server Panic / 500 Internal Server Error ─────────────────────────────────

func TestGetRate_ServerPanic_500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "api error") {
		t.Errorf("expected 'api error' in error, got: %v", err)
	}
}

// ─── Empty Body ───────────────────────────────────────────────────────────────

func TestGetRate_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write nothing — empty body
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected decode error for empty body, got nil")
	}
	if !strings.Contains(err.Error(), "decode error") {
		t.Errorf("expected 'decode error' in error, got: %v", err)
	}
}
