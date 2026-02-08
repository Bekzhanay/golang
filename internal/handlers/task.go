package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type TaskStore struct {
	mu     sync.Mutex
	nextID int
	tasks  map[int]Task
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		nextID: 1,
		tasks:  make(map[int]Task),
	}
}

type TaskHandler struct {
	store *TaskStore
}

func NewTaskHandler(store *TaskStore) *TaskHandler {
	return &TaskHandler{store: store}
}

func (h *TaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodPatch:
		h.handlePatch(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *TaskHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	if idStr == "" {
		h.store.mu.Lock()
		defer h.store.mu.Unlock()

		all := make([]Task, 0, len(h.store.tasks))
		for _, t := range h.store.tasks {
			all = append(all, t)
		}
		writeJSON(w, http.StatusOK, all)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	task, ok := h.store.tasks[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	type createTaskRequest struct {
		Title string `json:"title"`
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid title"})
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	task := Task{
		ID:    h.store.nextID,
		Title: title,
		Done:  false,
	}
	h.store.tasks[task.ID] = task
	h.store.nextID++

	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	type patchTaskRequest struct {
		Done *bool `json:"done"`
	}

	var req patchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if req.Done == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "done must be boolean"})
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	task, ok := h.store.tasks[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	task.Done = *req.Done
	h.store.tasks[id] = task

	writeJSON(w, http.StatusOK, map[string]any{
		"updated": true,
		"task":    task,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
