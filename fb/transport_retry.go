package fb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cenk/backoff"
)

type retryTransport struct {
	initialInterval time.Duration
	maxElapsedTime  time.Duration
	next            http.RoundTripper
}

func newRetryTransport(initialInterval, maxElapsedTime time.Duration, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	if initialInterval <= 0 {
		initialInterval = 5 * time.Millisecond
	}
	if maxElapsedTime <= 0 {
		maxElapsedTime = 1 * time.Minute
	}

	return &retryTransport{
		next:            next,
		initialInterval: initialInterval,
		maxElapsedTime:  maxElapsedTime,
	}
}

func (t *retryTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = t.initialInterval
	bo.MaxElapsedTime = t.maxElapsedTime
	var resp *http.Response
	var attempt int
	err := backoff.Retry(func() error {
		attempt++
		var e error

		resp, e = t.next.RoundTrip(r) // nolint:bodyclose // not a correct linter detection

		if e != nil {
			return e
		} else if resp.StatusCode >= 500 {
			defer resp.Body.Close()
			return retryStatusError(resp, attempt)
		}

		return nil
	}, backoff.WithContext(bo, r.Context()))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func retryStatusError(resp *http.Response, attempt int) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unexpected status %s from facebook, attempt %d: reading error: %v", resp.Status, attempt, err)
	}
	var errResp ErrorContainer
	if json.Unmarshal(body, &errResp) == nil {
		if fbErr := errResp.GetError(); fbErr != nil {
			if IsReduceData(fbErr) {
				return backoff.Permanent(fbErr)
			}

			return fmt.Errorf("unexpected status %s from facebook, attempt %d: %v", resp.Status, attempt, fbErr)
		}
	}

	return fmt.Errorf("unexpected status %s from facebook, attempt %d", resp.Status, attempt)
}
