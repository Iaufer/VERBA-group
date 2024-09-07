package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	DB *sql.DB
}

type Task struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Response struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodPut:
		h.handlePut(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/tasks") {
		idStr := strings.TrimPrefix(path, "/tasks")
		if idStr == "" {
			h.getAllTasks(w, r)
			return
		}
		id, err := strconv.Atoi(idStr[1:])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		err = h.getTaskById(w, r, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
	} else {
		http.NotFound(w, r)
	}
}

func (h *Handler) handlePost(w http.ResponseWriter, r *http.Request) {
	var t Task
	if err := decodeJSONBody(w, r, &t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateTask(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = t.CreatedAt

	id, err := h.insertTaskToDB(t)
	if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Server encountered an issue", http.StatusInternalServerError)
		return
	}

	t.Id = id
	respondWithJSON(w, http.StatusCreated, t)
}

func (h *Handler) handlePut(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var t Task
	if err := decodeJSONBody(w, r, &t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateTask(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t.UpdatedAt = time.Now().UTC()

	if err := h.updateTaskInDB(id, t); err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Server encountered an issue", http.StatusInternalServerError)
		return
	}

	t.Id = id
	respondWithJSON(w, http.StatusOK, t)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.deleteTaskFromDB(id); err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.fetchAllTasksFromDB()
	if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Server encountered an issue", http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, tasks)
}

func (h *Handler) getTaskById(w http.ResponseWriter, r *http.Request, id int) error {
	t, err := h.fetchTaskFromDB(id)
	if err != nil {
		return err
	}
	respondWithJSON(w, http.StatusOK, t)
	return nil
}

// Utility Functions

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading request body: %v", err)
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	return nil
}

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("Error encoding response:", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func validateTask(t *Task) error {
	if t.Title == "" || t.Description == "" {
		return fmt.Errorf("missing required fields: title and description")
	}

	if !t.DueDate.IsZero() && t.DueDate.Before(time.Now().UTC()) {
		return fmt.Errorf("invalid due_date: it cannot be in the past")
	}

	if t.DueDate.IsZero() {
		t.DueDate = time.Now().UTC().AddDate(0, 0, 7) // если латы нет то по дефолту + 7 дней от текушей даты
	}
	return nil
}

func extractIDFromPath(path string) (int, error) {
	idStr := strings.TrimPrefix(path, "/tasks/")
	if idStr == "" {
		return 0, fmt.Errorf("ID not provided")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format")
	}

	return id, nil
}

func (h *Handler) insertTaskToDB(t Task) (int, error) {
	var id int
	err := h.DB.QueryRow("INSERT INTO tasks (title, description, due_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		t.Title, t.Description, t.DueDate, t.CreatedAt, t.UpdatedAt).Scan(&id)
	return id, err
}

func (h *Handler) fetchTaskFromDB(id int) (Task, error) {
	var t Task
	err := h.DB.QueryRow("SELECT id, title, description, due_date, created_at, updated_at FROM tasks WHERE id=$1",
		id).Scan(&t.Id, &t.Title, &t.Description, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (h *Handler) fetchAllTasksFromDB() ([]Task, error) {
	rows, err := h.DB.Query("SELECT id, title, description, due_date, created_at, updated_at FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.Id, &t.Title, &t.Description, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (h *Handler) updateTaskInDB(id int, t Task) error {
	_, err := h.DB.Exec("UPDATE tasks SET title=$1, description=$2, due_date=$3, updated_at=$4 WHERE id=$5",
		t.Title, t.Description, t.DueDate, t.UpdatedAt, id)
	return err
}

func (h *Handler) deleteTaskFromDB(id int) error {
	result, err := h.DB.Exec("DELETE FROM tasks WHERE id=$1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}
