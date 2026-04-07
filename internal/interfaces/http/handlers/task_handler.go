package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"skoolz/database/postgres/repositories"
	"skoolz/internal/infrastructure/container"
	"skoolz/internal/shared/response"

	"github.com/google/uuid"
)

// TaskHandler handles task CRUD endpoints
type TaskHandler struct {
	repo *repositories.TaskRepository
}

// NewTaskHandler creates a new task handler using the global container's DB
func NewTaskHandler() *TaskHandler {
	db := container.GetContainer().GetDB()
	return &TaskHandler{repo: repositories.NewTaskRepository(db)}
}

// taskRequest is the payload for create/update
type taskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

func (req *taskRequest) toEntity(t *repositories.Task) {
	t.Title = req.Title
	t.Description = req.Description
	if req.Status != "" {
		t.Status = req.Status
	}
	if req.DueDate != nil {
		t.DueDate = sql.NullTime{Time: *req.DueDate, Valid: true}
	} else {
		t.DueDate = sql.NullTime{Valid: false}
	}
}

// Create POST /api/v1/tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteBadRequest(w, "Invalid JSON body")
		return
	}
	if req.Title == "" {
		response.WriteBadRequest(w, "title is required")
		return
	}

	task := &repositories.Task{}
	req.toEntity(task)

	if err := h.repo.Create(r.Context(), task); err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}
	response.WriteCreated(w, "Task created", task)
}

// List GET /api/v1/tasks?limit=&offset=
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	tasks, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}
	total, err := h.repo.Count(r.Context())
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}

	page := offset/limit + 1
	response.WritePaginated(w, http.StatusOK, "Tasks fetched", tasks, page, limit, total)
}

// Get GET /api/v1/tasks/{id}
func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}
	task, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}
	if task == nil {
		response.WriteNotFound(w, "Task not found")
		return
	}
	response.WriteOK(w, "Task fetched", task)
}

// Update PUT /api/v1/tasks/{id}
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}

	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}
	if existing == nil {
		response.WriteNotFound(w, "Task not found")
		return
	}

	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteBadRequest(w, "Invalid JSON body")
		return
	}
	req.toEntity(existing)

	if err := h.repo.Update(r.Context(), existing); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteNotFound(w, "Task not found")
			return
		}
		response.WriteInternalServerError(w, err.Error())
		return
	}
	response.WriteOK(w, "Task updated", existing)
}

// Delete DELETE /api/v1/tasks/{id}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteNotFound(w, "Task not found")
			return
		}
		response.WriteInternalServerError(w, err.Error())
		return
	}
	response.WriteOK(w, "Task deleted", nil)
}

func parseTaskID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WriteBadRequest(w, "Invalid task id")
		return uuid.Nil, false
	}
	return id, true
}
