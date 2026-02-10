package api

import (
	"net/http"

	"github.com/LanaAntonova/go-final-proj/pkg/db"
)

var limit = 50

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.Tasks(limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteJSON(w, map[string]string{"error": "Failed to receive task: " + err.Error()})
		return
	}

	if tasks == nil {
		tasks = []*db.Task{}
	}

	WriteJSON(w, TasksResp{Tasks: tasks})
}
