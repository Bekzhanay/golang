package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"time"

	"practice9/idempotency"
)

func main() {
	store := idempotency.NewMemoryStore()

	var businessLogicCount int32

	paymentHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&businessLogicCount, 1)

		fmt.Printf("Processing started. Business logic execution #%d\n", count)

		time.Sleep(2 * time.Second)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := fmt.Sprintf(
			`{"status":"paid","amount":1000,"transaction_id":"uuid-%d"}`,
			count,
		)

		w.Write([]byte(response))

		fmt.Printf("Processing finished. Response: %s\n", response)
	})

	protectedHandler := idempotency.Middleware(store, paymentHandler)

	server := httptest.NewServer(protectedHandler)
	defer server.Close()

	key := "same-key-123"

	fmt.Println("=== Sending 10 simultaneous requests with the same Idempotency-Key ===")

	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)

		go func(requestID int) {
			defer wg.Done()

			req, err := http.NewRequest(http.MethodPost, server.URL+"/pay", bytes.NewBuffer(nil))
			if err != nil {
				fmt.Printf("Request %d: error creating request: %v\n", requestID, err)
				return
			}

			req.Header.Set("Idempotency-Key", key)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("Request %d: client error: %v\n", requestID, err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			fmt.Printf("Request %d: status=%d body=%s\n", requestID, resp.StatusCode, string(body))
		}(i)
	}

	wg.Wait()

	fmt.Println()
	fmt.Println("=== Sending one more request after completion with the same key ===")

	req, err := http.NewRequest(http.MethodPost, server.URL+"/pay", bytes.NewBuffer(nil))
	if err != nil {
		fmt.Println("error creating final request:", err)
		return
	}

	req.Header.Set("Idempotency-Key", key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("final request error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Final repeated request: status=%d body=%s\n", resp.StatusCode, string(body))

	fmt.Println()
	fmt.Printf("Business logic was executed only %d time(s)\n", atomic.LoadInt32(&businessLogicCount))
}
