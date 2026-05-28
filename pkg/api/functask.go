package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LanaAntonova/go-final-proj/pkg/db"
)

func WriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(data)
}

func isValidDate(date string) bool {
	for _, c := range date {
		if c < '0' || c > '9' {
			return false
		}
	}
	year, _ := strconv.Atoi(date[0:4])
	month, _ := strconv.Atoi(date[4:6])
	day, _ := strconv.Atoi(date[6:8])
	return year >= 1000 && year <= 3000 && month >= 1 && month <= 12 && day >= 1 && day <= 31
}

func isValidRepeat(repeat string) bool {
	parts := strings.Fields(repeat)
	if len(parts) != 2 {
		return false
	}
	period := parts[0]
	count, err := strconv.Atoi(parts[1])
	if err != nil || count <= 0 {
		return false
	}
	switch period {
	case "d", "w", "m", "y":
		return true
	default:
		return false
	}
}

func afterNow(date time.Time, now time.Time) bool {
	return date.Format(Layout) >= now.Format(Layout)
}

func checkDate(task *db.Task, now time.Time) error {
	if task.Date == "" {
		task.Date = now.Format(Layout)
		return nil
	}

	_, err := time.Parse(Layout, task.Date)
	if err != nil {
		return errors.New("Invalid date format, expected YYYYMMDD")
	}

	if task.Repeat != "" {
		_, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	return nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(task.Title) == "" {
		WriteJSON(w, map[string]string{"error": "The task name is not specified"})
		return
	}

	now := time.Now()

	if err := checkDate(&task, now); err != nil {
		WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}

	t, _ := time.Parse(Layout, task.Date)

	if !afterNow(t, now) {
		if task.Repeat == "" {
			task.Date = now.Format(Layout)
		} else {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				WriteJSON(w, map[string]string{"error": err.Error()})
				return
			}
			task.Date = next
		}
	}

	id, err := db.AddTask(&task)
	if err != nil {
		WriteJSON(w, map[string]string{"error": "Failed to save task: " + err.Error()})
		return
	}

	WriteJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		WriteJSON(w, map[string]string{"error": "Task ID is not specified"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		WriteJSON(w, map[string]string{"error": "Failed to get task: " + err.Error()})
		return
	}

	WriteJSON(w, task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		WriteJSON(w, map[string]string{"error": "Invalid JSON"})
		return
	}

	if task.Title == "" {
		WriteJSON(w, map[string]string{"error": "The task name is not specified"})
		return
	}
	if len(task.Date) != 8 || !isValidDate(task.Date) {
		WriteJSON(w, map[string]string{"error": "Invalid date format"})
		return
	}
	if task.Repeat != "" && !isValidRepeat(task.Repeat) {
		WriteJSON(w, map[string]string{"error": "Invalid repeat rule"})
		return
	}

	err := db.UpdateTask(&task)
	if err != nil {
		WriteJSON(w, map[string]string{"error": "Failed to update task: " + err.Error()})
		return
	}

	WriteJSON(w, map[string]interface{}{})
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		WriteJSON(w, map[string]string{"error": "Task ID is not specified"})
		return
	}

	err := db.DeleteTask(id)
	if err != nil {
		WriteJSON(w, map[string]string{"error": "Failed to delete task: " + err.Error()})
		return
	}

	WriteJSON(w, map[string]interface{}{})
	//w.WriteHeader(http.StatusNoContent)
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		addTaskHandler(w, r)
	case "GET":
		getTaskHandler(w, r)
	case "PUT":
		updateTaskHandler(w, r)
	case "DELETE":
		deleteTaskHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
