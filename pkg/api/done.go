package api

import (
	"net/http"
	"time"

	"github.com/LanaAntonova/go-final-proj/pkg/db"
)

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			WriteJSON(w, map[string]string{"error": "Failed to delete task: " + err.Error()})
			return
		}
	} else {
		now := time.Now()
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			WriteJSON(w, map[string]string{"error": err.Error()})
			return
		}
		err = db.UpdateDate(next, id)
		if err != nil {
			WriteJSON(w, map[string]string{"error": "Failed to update task date: " + err.Error()})
			return
		}
	}

	WriteJSON(w, map[string]interface{}{})
}
