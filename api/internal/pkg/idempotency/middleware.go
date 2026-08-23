package idempotency

import (
	"bytes"
	"net/http"
	"sync"
	"time"
)

type storedResponse struct {
	statusCode int
	body       []byte
	header     http.Header
	timestamp  time.Time
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]storedResponse
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items: make(map[string]storedResponse),
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func Middleware(store *MemoryStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" || (r.Method != http.MethodPost && r.Method != http.MethodPut) {
				next.ServeHTTP(w, r)
				return
			}

			store.mu.RLock()
			cached, exists := store.items[key]
			store.mu.RUnlock()

			if exists {
				for k, v := range cached.header {
					w.Header()[k] = v
				}
				w.Header().Set("X-Cache", "HIT-Idempotent")
				w.WriteHeader(cached.statusCode)
				_, _ = w.Write(cached.body)
				return
			}

			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			if rec.statusCode >= 200 && rec.statusCode < 300 {
				store.mu.Lock()
				store.items[key] = storedResponse{
					statusCode: rec.statusCode,
					body:       rec.body.Bytes(),
					header:     rec.Header().Clone(),
					timestamp:  time.Now(),
				}
				store.mu.Unlock()
			}
		})
	}
}
