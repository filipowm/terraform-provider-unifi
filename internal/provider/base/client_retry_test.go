package base

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newTestRetryRoundTripper builds a retryRoundTripper with a near-zero backoff so
// tests do not sleep.
func newTestRetryRoundTripper(next http.RoundTripper, maxRetries int) *retryRoundTripper {
	return &retryRoundTripper{
		next:       next,
		maxRetries: maxRetries,
		backoff:    time.Millisecond,
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	if resp == nil || resp.Body == nil {
		return ""
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	return string(data)
}

func TestRetryRoundTripper_RetriesOn5xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	rt := newTestRetryRoundTripper(http.DefaultTransport, 5)
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if body := readBody(t, resp); body != `{"ok":true}` {
		t.Fatalf("unexpected body: %q", body)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestRetryRoundTripper_RetriesOnHTMLBody(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 2 {
			// HTML body with a 200 status and no text/html content type to exercise
			// the body-sniffing path.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("\n  <html><body>nginx</body></html>"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	rt := newTestRetryRoundTripper(http.DefaultTransport, 3)
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body := readBody(t, resp); body != `{"ok":true}` {
		t.Fatalf("unexpected body: %q", body)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls, got %d", got)
	}
}

func TestRetryRoundTripper_PreservesHTMLBodyWhenExhausted(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html>still broken</html>"))
	}))
	defer srv.Close()

	rt := newTestRetryRoundTripper(http.DefaultTransport, 2)
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Body must still be fully readable after retries are exhausted.
	if body := readBody(t, resp); !strings.Contains(body, "still broken") {
		t.Fatalf("expected html body returned, got %q", body)
	}
	if got := atomic.LoadInt32(&calls); got != 3 { // initial + 2 retries
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestRetryRoundTripper_DoesNotRetryPOST(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()

	rt := newTestRetryRoundTripper(http.DefaultTransport, 5)
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, srv.URL, strings.NewReader(`{"a":1}`))
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = readBody(t, resp)
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected POST to be attempted once, got %d", got)
	}
}

func TestRetryRoundTripper_StopsAfterMaxRetries(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()

	rt := newTestRetryRoundTripper(http.DefaultTransport, 2)
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = readBody(t, resp)
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 3 { // initial + 2 retries
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestRetryRoundTripper_RetriesPUTWithBody(t *testing.T) {
	var calls int32
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		data, _ := io.ReadAll(r.Body)
		lastBody = string(data)
		if n < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	rt := newTestRetryRoundTripper(http.DefaultTransport, 3)
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPut, srv.URL, strings.NewReader(`{"a":1}`))
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = readBody(t, resp)
	if lastBody != `{"a":1}` {
		t.Fatalf("expected body to be replayed, got %q", lastBody)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls, got %d", got)
	}
}

func TestNewRetryRoundTripper_PassThroughWhenDisabled(t *testing.T) {
	next := http.DefaultTransport
	got := newRetryRoundTripper(next, 0)
	if _, ok := got.(*retryRoundTripper); ok {
		t.Fatalf("expected pass-through transport when maxRetries==0, got wrapper")
	}

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	resp, err := got.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = readBody(t, resp)
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 call with retries disabled, got %d", got)
	}
}
