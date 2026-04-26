package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"practice9/retry"
)

func main() {
	var counter int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&counter, 1)

		if count <= 3 {
			fmt.Printf("Payment Gateway: request %d -> 503 Service Unavailable\n", count)
			http.Error(w, "temporary failure", http.StatusServiceUnavailable)
			return
		}

		fmt.Printf("Payment Gateway: request %d -> 200 OK\n", count)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := retry.PaymentClient{
		Client:  &http.Client{Timeout: 3 * time.Second},
		BaseURL: server.URL,
		Config: retry.Config{
			MaxRetries: 5,
			BaseDelay:  500 * time.Millisecond,
			MaxDelay:   5 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.ExecutePayment(ctx)
	if err != nil {
		fmt.Println("Final result: failed:", err)
		return
	}

	fmt.Println("Final result: payment succeeded")
}
