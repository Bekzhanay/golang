package idempotency

import (
	"net/http"
	"net/http/httptest"
	"sync"
)

type CachedResponse struct {
	StatusCode int
	Body       []byte
	Completed  bool
}

type MemoryStore struct {
	mu   sync.Mutex
	data map[string]*CachedResponse
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]*CachedResponse),
	}
}

func (m *MemoryStore) Get(key string) (*CachedResponse, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	resp, exists := m.data[key]
	return resp, exists
}

func (m *MemoryStore) StartProcessing(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[key]; exists {
		return false
	}

	m.data[key] = &CachedResponse{
		Completed: false,
	}

	return true
}

func (m *MemoryStore) Finish(key string, status int, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if resp, exists := m.data[key]; exists {
		resp.StatusCode = status
		resp.Body = body
		resp.Completed = true
		return
	}

	m.data[key] = &CachedResponse{
		StatusCode: status,
		Body:       body,
		Completed:  true,
	}
}

func Middleware(store *MemoryStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
			return
		}

		if cached, exists := store.Get(key); exists {
			if cached.Completed {
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
				return
			}

			http.Error(w, "Duplicate request in progress", http.StatusConflict)
			return
		}

		if !store.StartProcessing(key) {
			if cached, exists := store.Get(key); exists && cached.Completed {
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
				return
			}

			http.Error(w, "Duplicate request in progress", http.StatusConflict)
			return
		}

		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)

		body := recorder.Body.Bytes()
		store.Finish(key, recorder.Code, body)

		for k, vals := range recorder.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}

		w.WriteHeader(recorder.Code)
		w.Write(body)
	})
}