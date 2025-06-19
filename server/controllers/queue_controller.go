package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// QueueController handles all requests relating to Atlantis plan queues.
type QueueController struct {
	AtlantisVersion    string                       `validate:"required"`
	AtlantisURL        *url.URL                     `validate:"required"`
	Logger             logging.SimpleLogging        `validate:"required"`
	QueueManager       models.PlanQueueManager      `validate:"required"`
	QueueTemplate      web_templates.TemplateWriter `validate:"required"`
}

// QueueData represents the data structure for queue template rendering
type QueueData struct {
	Queues         []QueueInfo
	AtlantisVersion string
	CleanedBasePath string
}

// QueueInfo represents a single queue entry for the UI
type QueueInfo struct {
	Project     string
	Workspace   string
	RepoFullName string
	Entries     []QueueEntryInfo
	UpdatedAt   string
}

// QueueEntryInfo represents a single queue entry for the UI
type QueueEntryInfo struct {
	ID       string
	PullNum  int
	Username string
	Time     string
	Position int
}

// GetQueues is the GET /queues route. It renders the queue overview page.
func (q *QueueController) GetQueues(w http.ResponseWriter, r *http.Request) {
	// Get all active queues
	queues, err := q.QueueManager.GetAllQueues()
	if err != nil {
		q.respond(w, logging.Error, http.StatusInternalServerError, "Failed getting queues: %s", err)
		return
	}

	// Convert to UI-friendly format
	queueInfos := make([]QueueInfo, 0, len(queues))
	for _, queue := range queues {
		entries := make([]QueueEntryInfo, 0, len(queue.Entries))
		for i, entry := range queue.Entries {
			entries = append(entries, QueueEntryInfo{
				ID:       entry.ID,
				PullNum:  entry.Pull.Num,
				Username: entry.User.Username,
				Time:     entry.Time.Format("2006-01-02 15:04:05"),
				Position: i + 1,
			})
		}

		queueInfos = append(queueInfos, QueueInfo{
			Project:     queue.Project.String(),
			Workspace:   queue.Workspace,
			RepoFullName: queue.Project.RepoFullName,
			Entries:     entries,
			UpdatedAt:   queue.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	viewData := QueueData{
		Queues:         queueInfos,
		AtlantisVersion: q.AtlantisVersion,
		CleanedBasePath: q.AtlantisURL.Path,
	}

	err = q.QueueTemplate.Execute(w, viewData)
	if err != nil {
		q.Logger.Err(err.Error())
	}
}

// GetQueueStatus is the GET /api/queues/{repo}/{project}/{workspace} route.
// It returns JSON data for a specific queue.
func (q *QueueController) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoFullName := vars["repo"]
	projectPath := vars["project"]
	workspace := vars["workspace"]

	project := models.Project{
		RepoFullName: repoFullName,
		Path:         projectPath,
	}

	queue, err := q.QueueManager.GetQueueStatus(project, workspace)
	if err != nil {
		q.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed getting queue status: %s", err)})
		return
	}

	if queue == nil {
		q.respondJSON(w, http.StatusNotFound, map[string]string{"error": "Queue not found"})
		return
	}

	// Convert to JSON-friendly format
	entries := make([]map[string]interface{}, 0, len(queue.Entries))
	for i, entry := range queue.Entries {
		entries = append(entries, map[string]interface{}{
			"id":       entry.ID,
			"pull_num": entry.Pull.Num,
			"username": entry.User.Username,
			"time":     entry.Time.Format("2006-01-02 15:04:05"),
			"position": i + 1,
		})
	}

	response := map[string]interface{}{
		"project":   queue.Project.String(),
		"workspace": queue.Workspace,
		"entries":   entries,
		"updated_at": queue.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	q.respondJSON(w, http.StatusOK, response)
}

// RemoveFromQueue is the DELETE /api/queues/{repo}/{project}/{workspace}/{pull_num} route.
// It removes a specific entry from the queue.
func (q *QueueController) RemoveFromQueue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoFullName := vars["repo"]
	projectPath := vars["project"]
	workspace := vars["workspace"]
	pullNum := vars["pull_num"]

	// Parse pull number
	var pullNumInt int
	_, err := fmt.Sscanf(pullNum, "%d", &pullNumInt)
	if err != nil {
		q.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid pull number"})
		return
	}

	project := models.Project{
		RepoFullName: repoFullName,
		Path:         projectPath,
	}

	err = q.QueueManager.RemoveFromQueue(project, workspace, pullNumInt)
	if err != nil {
		q.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed removing from queue: %s", err)})
		return
	}

	q.respondJSON(w, http.StatusOK, map[string]string{"message": "Successfully removed from queue"})
}

// GetAllQueues is the GET /api/queues route.
// It returns JSON data for all active queues.
func (q *QueueController) GetAllQueues(w http.ResponseWriter, r *http.Request) {
	queues, err := q.QueueManager.GetAllQueues()
	if err != nil {
		q.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed getting queues: %s", err)})
		return
	}

	// Convert to JSON-friendly format
	queueList := make([]map[string]interface{}, 0, len(queues))
	for _, queue := range queues {
		entries := make([]map[string]interface{}, 0, len(queue.Entries))
		for i, entry := range queue.Entries {
			entries = append(entries, map[string]interface{}{
				"id":       entry.ID,
				"pull_num": entry.Pull.Num,
				"username": entry.User.Username,
				"time":     entry.Time.Format("2006-01-02 15:04:05"),
				"position": i + 1,
			})
		}

		queueList = append(queueList, map[string]interface{}{
			"project":   queue.Project.String(),
			"workspace": queue.Workspace,
			"repo_full_name": queue.Project.RepoFullName,
			"entries":   entries,
			"updated_at": queue.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	q.respondJSON(w, http.StatusOK, map[string]interface{}{
		"queues": queueList,
		"count":  len(queueList),
	})
}

// respond is a helper function to respond and log the response.
func (q *QueueController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	q.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}

// respondJSON is a helper function to respond with JSON data.
func (q *QueueController) respondJSON(w http.ResponseWriter, responseCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		q.Logger.Err("Failed to marshal JSON response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	w.Write(jsonData)
} 