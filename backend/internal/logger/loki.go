package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LokiConfig struct {
	URL           string
	User          string
	Token         string
	Labels        map[string]string
	BatchInterval time.Duration
	BatchSize     int
	MaxBufferSize int
	HTTPClient    *http.Client
	// retryBackoffs lets tests shrink the retry schedule. nil → production defaults.
	retryBackoffs []time.Duration
}

type LokiHandler struct {
	inner  slog.Handler
	shared *lokiShared
}

type lokiEntry struct {
	ts   time.Time
	line string
}

type lokiShared struct {
	pushURL       string
	user          string
	token         string
	labels        map[string]string
	batchSize     int
	maxBuffer     int
	interval      time.Duration
	retryBackoffs []time.Duration
	httpClient    *http.Client

	mu      sync.Mutex
	buf     []lokiEntry
	flushCh chan struct{}
	doneCh  chan struct{}
	closed  bool
	wg      sync.WaitGroup
}

func NewLokiHandler(cfg LokiConfig) *LokiHandler {
	if cfg.BatchInterval <= 0 {
		cfg.BatchInterval = 5 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	if cfg.MaxBufferSize <= 0 {
		cfg.MaxBufferSize = 10000
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	backoffs := cfg.retryBackoffs
	if backoffs == nil {
		backoffs = []time.Duration{0, 200 * time.Millisecond, 500 * time.Millisecond, 2 * time.Second}
	}

	shared := &lokiShared{
		pushURL:       normalizeURL(cfg.URL),
		user:          cfg.User,
		token:         cfg.Token,
		labels:        cfg.Labels,
		batchSize:     cfg.BatchSize,
		maxBuffer:     cfg.MaxBufferSize,
		interval:      cfg.BatchInterval,
		retryBackoffs: backoffs,
		httpClient:    cfg.HTTPClient,
		buf:           make([]lokiEntry, 0, cfg.BatchSize),
		flushCh:       make(chan struct{}, 1),
		doneCh:        make(chan struct{}),
	}

	capturer := &lineCapturer{shared: shared}
	inner := slog.NewJSONHandler(capturer, &slog.HandlerOptions{Level: slog.LevelInfo})

	shared.wg.Add(1)
	go shared.run()
	return &LokiHandler{inner: inner, shared: shared}
}

func normalizeURL(u string) string {
	u = strings.TrimRight(u, "/")
	if strings.HasSuffix(u, "/loki/api/v1/push") {
		return u
	}
	return u + "/loki/api/v1/push"
}

type lineCapturer struct {
	shared *lokiShared
}

func (lc *lineCapturer) Write(p []byte) (int, error) {
	line := strings.TrimRight(string(p), "\n")
	lc.shared.append(time.Now(), line)
	return len(p), nil
}

func (h *LokiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *LokiHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.inner.Handle(ctx, r)
}

func (h *LokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LokiHandler{inner: h.inner.WithAttrs(attrs), shared: h.shared}
}

func (h *LokiHandler) WithGroup(name string) slog.Handler {
	return &LokiHandler{inner: h.inner.WithGroup(name), shared: h.shared}
}

func (h *LokiHandler) Close(ctx context.Context) error {
	h.shared.mu.Lock()
	if h.shared.closed {
		h.shared.mu.Unlock()
		return nil
	}
	h.shared.closed = true
	h.shared.mu.Unlock()

	close(h.shared.doneCh)

	done := make(chan struct{})
	go func() {
		h.shared.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *lokiShared) append(ts time.Time, line string) {
	s.mu.Lock()
	if len(s.buf) >= s.maxBuffer {
		// drop oldest — newer events are more useful for current debugging
		copy(s.buf, s.buf[1:])
		s.buf = s.buf[:len(s.buf)-1]
	}
	s.buf = append(s.buf, lokiEntry{ts: ts, line: line})
	shouldFlush := len(s.buf) >= s.batchSize
	s.mu.Unlock()

	if shouldFlush {
		select {
		case s.flushCh <- struct{}{}:
		default:
		}
	}
}

func (s *lokiShared) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.doneCh:
			s.flush()
			return
		case <-ticker.C:
			s.flush()
		case <-s.flushCh:
			s.flush()
		}
	}
}

func (s *lokiShared) flush() {
	s.mu.Lock()
	if len(s.buf) == 0 {
		s.mu.Unlock()
		return
	}
	batch := s.buf
	s.buf = make([]lokiEntry, 0, s.batchSize)
	s.mu.Unlock()

	payload, err := buildPayload(s.labels, batch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loki: dropped %d records: build payload: %v\n", len(batch), err)
		return
	}
	s.push(payload, len(batch))
}

func buildPayload(labels map[string]string, entries []lokiEntry) ([]byte, error) {
	sortEntriesAscending(entries)
	for i := 1; i < len(entries); i++ {
		if !entries[i].ts.After(entries[i-1].ts) {
			entries[i].ts = entries[i-1].ts.Add(time.Nanosecond)
		}
	}

	values := make([][2]string, len(entries))
	for i, e := range entries {
		values[i] = [2]string{strconv.FormatInt(e.ts.UnixNano(), 10), e.line}
	}

	body := map[string]any{
		"streams": []map[string]any{{
			"stream": labels,
			"values": values,
		}},
	}
	return json.Marshal(body)
}

func sortEntriesAscending(entries []lokiEntry) {
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].ts.Before(entries[j-1].ts); j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}
}

func (s *lokiShared) push(payload []byte, count int) {
	var lastErr error
	for _, backoff := range s.retryBackoffs {
		if backoff > 0 {
			time.Sleep(backoff)
		}

		req, err := http.NewRequest("POST", s.pushURL, bytes.NewReader(payload))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(s.user, s.token)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// 4xx (except 429): bad input or auth — do not retry
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			fmt.Fprintf(os.Stderr, "loki: dropped %d records: HTTP %d (no retry)\n", count, resp.StatusCode)
			return
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			return
		}

		// 5xx or 429: retry
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	fmt.Fprintf(os.Stderr, "loki: dropped %d records: %v\n", count, lastErr)
}
