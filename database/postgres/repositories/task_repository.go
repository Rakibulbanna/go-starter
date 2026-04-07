package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Task represents a task entity
type Task struct {
	ID          uuid.UUID    `db:"id" json:"id"`
	Title       string       `db:"title" json:"title"`
	Description string       `db:"description" json:"description"`
	Status      string       `db:"status" json:"status"`
	DueDate     sql.NullTime `db:"due_date" json:"due_date"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
}

// TaskRepository handles database operations for tasks
type TaskRepository struct {
	db *sqlx.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *sqlx.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create inserts a new task
func (r *TaskRepository) Create(ctx context.Context, task *Task) error {
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	if task.Status == "" {
		task.Status = "pending"
	}

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	query := `
		INSERT INTO tasks (id, title, description, status, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.DueDate, task.CreatedAt, task.UpdatedAt)
	return err
}

// GetByID retrieves a task by its ID
func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	var task Task
	query := `SELECT * FROM tasks WHERE id = $1`
	err := r.db.GetContext(ctx, &task, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// List retrieves tasks with pagination
func (r *TaskRepository) List(ctx context.Context, limit, offset int) ([]*Task, error) {
	var tasks []*Task
	query := `SELECT * FROM tasks ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &tasks, query, limit, offset); err != nil {
		return nil, err
	}
	return tasks, nil
}

// Update updates a task
func (r *TaskRepository) Update(ctx context.Context, task *Task) error {
	task.UpdatedAt = time.Now()
	query := `
		UPDATE tasks
		SET title = $2, description = $3, status = $4, due_date = $5, updated_at = $6
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.DueDate, task.UpdatedAt)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Delete removes a task by ID
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Count returns the total number of tasks
func (r *TaskRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM tasks`)
	return count, err
}
