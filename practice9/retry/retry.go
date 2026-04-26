package retry

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"time"
)

type Config struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type PaymentClient struct {
	Client  *http.Client
	BaseURL string
	Config  Config
}

func IsRetryable(resp *http.Response, err error) bool {
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return true
		}
		return true
	}

	if resp == nil {
		return false
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	case http.StatusUnauthorized, http.StatusNotFound:
		return false
	default:
		return false
	}
}

func CalculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	backoff := baseDelay * (1 << attempt)
	if backoff > maxDelay {
		backoff = maxDelay
	}

	if backoff <= 0 {
		return 0
	}

	return time.Duration(rand.Int63n(int64(backoff)))
}

func (p *PaymentClient) ExecutePayment(ctx context.Context) error {
	for attempt := 0; attempt < p.Config.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/pay", nil)
		if err != nil {
			return err
		}

		fmt.Printf("Attempt %d: sending payment request...\n", attempt+1)

		resp, err := p.Client.Do(req)

		if err == nil && resp != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Printf("Attempt %d: Success! response=%s\n", attempt+1, string(body))
				return nil
			}

			fmt.Printf("Attempt %d failed with status %d\n", attempt+1, resp.StatusCode)
		} else {
			fmt.Printf("Attempt %d failed with error: %v\n", attempt+1, err)
		}

		if !IsRetryable(resp, err) {
			return fmt.Errorf("non-retryable error, stopping")
		}

		if attempt == p.Config.MaxRetries-1 {
			break
		}

		delay := CalculateBackoff(attempt, p.Config.BaseDelay, p.Config.MaxDelay)
		fmt.Printf("Attempt %d failed: waiting %v before retry...\n", attempt+1, delay)

		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return fmt.Errorf("payment failed after %d attempts", p.Config.MaxRetries)
}
