package logger

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNormalizeURL(t *testing.T) {
	cases := map[string]string{
		"https://logs-prod-006.grafana.net":                       "https://logs-prod-006.grafana.net/loki/api/v1/push",
		"https://logs-prod-006.grafana.net/":                      "https://logs-prod-006.grafana.net/loki/api/v1/push",
		"https://logs-prod-006.grafana.net/loki/api/v1/push":      "https://logs-prod-006.grafana.net/loki/api/v1/push",
		"https://logs-prod-006.grafana.net/loki/api/v1/push/":     "https://logs-prod-006.grafana.net/loki/api/v1/push",
		"http://localhost:3100":                                   "http://localhost:3100/loki/api/v1/push",
	}
	for in, want := range cases {
		if got := normalizeURL(in); got != want {
			t.Errorf("normalizeURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildPayloadSortsAndDedupes(t *testing.T) {
	now := time.Now()
	entries := []lokiEntry{
		{ts: now.Add(2 * time.Millisecond), line: "third"},
		{ts: now, line: "first"},
		{ts: now, line: "duplicate-of-first"},
		{ts: now.Add(1 * time.Millisecond), line: "second"},
	}

	payload, err := buildPayload(map[string]string{"service": "test"}, entries)
	if err != nil {
		t.Fatalf("buildPayload: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	streams := got["streams"].([]any)
	stream := streams[0].(map[string]any)
	values := stream["values"].([]any)

	if len(values) != 4 {
		t.Fatalf("want 4 values, got %d", len(values))
	}

	// Confirm order is first, dup, second, third
	wantOrder := []string{"first", "duplicate-of-first", "second", "third"}
	for i, want := range wantOrder {
		v := values[i].([]any)
		if v[1].(string) != want {
			t.Errorf("values[%d] = %v, want %q", i, v[1], want)
		}
	}

	// Confirm timestamps are strictly increasing (dedupe bumped duplicate by +1ns)
	var lastTs int64
	for i, vAny := range values {
		v := vAny.([]any)
		ts := v[0].(string)
		var n int64
		if _, err := jsonNumberToInt64(ts, &n); err != nil {
			t.Fatalf("ts parse: %v", err)
		}
		if i > 0 && n <= lastTs {
			t.Errorf("values[%d] ts %d not strictly > previous %d", i, n, lastTs)
		}
		lastTs = n
	}
}

func jsonNumberToInt64(s string, out *int64) (int64, error) {
	var n int64
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, &jsonParseErr{s: s}
		}
		n = n*10 + int64(r-'0')
	}
	*out = n
	return n, nil
}

type jsonParseErr struct{ s string }

func (e *jsonParseErr) Error() string { return "not a number: " + e.s }

func TestLokiHandlerHappyPath(t *testing.T) {
	var (
		mu       sync.Mutex
		received []byte
		auth     string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		received = body
		auth = r.Header.Get("Authorization")
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	h := NewLokiHandler(LokiConfig{
		URL:           server.URL,
		User:          "user1",
		Token:         "token1",
		Labels:        map[string]string{"service": "test", "env": "test"},
		BatchInterval: 30 * time.Millisecond,
		BatchSize:     100,
	})
	defer h.Close(context.Background())

	logger := slog.New(h)
	logger.Info("hello", "key", "value")

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		got := len(received)
		mu.Unlock()
		if got > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) == 0 {
		t.Fatal("no payload received within deadline")
	}

	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("user1:token1"))
	if auth != want {
		t.Errorf("auth: want %q, got %q", want, auth)
	}

	var got map[string]any
	if err := json.Unmarshal(received, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	streams := got["streams"].([]any)
	if len(streams) != 1 {
		t.Fatalf("want 1 stream, got %d", len(streams))
	}
	stream := streams[0].(map[string]any)
	labels := stream["stream"].(map[string]any)
	if labels["service"] != "test" {
		t.Errorf("service label = %v", labels["service"])
	}
	values := stream["values"].([]any)
	if len(values) != 1 {
		t.Fatalf("want 1 value, got %d", len(values))
	}
	v := values[0].([]any)
	line := v[1].(string)
	if !contains(line, `"msg":"hello"`) || !contains(line, `"key":"value"`) {
		t.Errorf("line missing expected fields: %s", line)
	}
}

func TestLokiHandlerNoRetryOn4xx(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	h := NewLokiHandler(LokiConfig{
		URL:           server.URL,
		User:          "u",
		Token:         "t",
		Labels:        map[string]string{"service": "test"},
		BatchInterval: 30 * time.Millisecond,
		retryBackoffs: []time.Duration{0, 5 * time.Millisecond, 5 * time.Millisecond, 5 * time.Millisecond},
	})

	slog.New(h).Info("x")

	time.Sleep(300 * time.Millisecond)
	h.Close(context.Background())

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("attempts: want 1 (no retry on 4xx), got %d", got)
	}
}

func TestLokiHandlerRetriesOn5xx(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	h := NewLokiHandler(LokiConfig{
		URL:           server.URL,
		User:          "u",
		Token:         "t",
		Labels:        map[string]string{"service": "test"},
		BatchInterval: 30 * time.Millisecond,
		retryBackoffs: []time.Duration{0, 5 * time.Millisecond, 10 * time.Millisecond, 20 * time.Millisecond},
	})

	slog.New(h).Info("x")

	time.Sleep(500 * time.Millisecond)
	h.Close(context.Background())

	if got := atomic.LoadInt32(&attempts); got != 4 {
		t.Errorf("attempts: want 4 (1 initial + 3 retries), got %d", got)
	}
}

func TestLokiHandlerBufferOverflowDropsOldest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	h := NewLokiHandler(LokiConfig{
		URL:           server.URL,
		User:          "u",
		Token:         "t",
		Labels:        map[string]string{"service": "test"},
		BatchInterval: 1 * time.Hour, // never auto-flush
		BatchSize:     1_000_000,     // never threshold-flush
		MaxBufferSize: 10,
	})
	defer h.Close(context.Background())

	logger := slog.New(h)
	for i := 0; i < 30; i++ {
		logger.Info("x", "i", i)
	}

	h.shared.mu.Lock()
	got := len(h.shared.buf)
	h.shared.mu.Unlock()

	if got != 10 {
		t.Errorf("buffer length = %d, want 10 (oldest dropped)", got)
	}
}

func TestFanoutClonesRecord(t *testing.T) {
	a := &capturingHandler{}
	b := &capturingHandler{}
	logger := slog.New(NewFanout(a, b))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			logger.Info("msg", "i", i)
		}(i)
	}
	wg.Wait()

	if a.count() != 100 || b.count() != 100 {
		t.Errorf("a=%d b=%d, want both 100", a.count(), b.count())
	}
}

type capturingHandler struct {
	mu sync.Mutex
	n  int
}

func (c *capturingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (c *capturingHandler) Handle(_ context.Context, r slog.Record) error {
	r.Attrs(func(slog.Attr) bool { return true })
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
	return nil
}
func (c *capturingHandler) WithAttrs([]slog.Attr) slog.Handler { return c }
func (c *capturingHandler) WithGroup(string) slog.Handler      { return c }

func (c *capturingHandler) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
