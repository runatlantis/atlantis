package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/internal/application/commands"
	"github.com/runatlantis/atlantis/internal/domain/project"
)

// ProjectHandler handles HTTP requests for project operations
type ProjectHandler struct {
	planHandler *commands.PlanProjectHandler
}

func NewProjectHandler(planHandler *commands.PlanProjectHandler) *ProjectHandler {
	return &ProjectHandler{
		planHandler: planHandler,
	}
}

func (h *ProjectHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/projects/{id}/plan", h.PlanProject).Methods("POST")
}

// PlanProjectRequest represents the HTTP request body for planning a project
type PlanProjectRequest struct {
	UserID      string `json:"user_id"`
	PullRequest struct {
		Number     int    `json:"number"`
		Repository string `json:"repository"`
		HeadSHA    string `json:"head_sha"`
	} `json:"pull_request"`
}

// PlanProjectResponse represents the HTTP response for planning a project
type PlanProjectResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	PlanFile string `json:"plan_file,omitempty"`
	Output   string `json:"output,omitempty"`
}

func (h *ProjectHandler) PlanProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := project.ProjectID(vars["id"])

	var req PlanProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert HTTP request to application command
	cmd := commands.PlanProjectCommand{
		ProjectID: projectID,
		UserID:    req.UserID,
		PullRequest: commands.PullRequestInfo{
			Number:     req.PullRequest.Number,
			Repository: req.PullRequest.Repository,
			HeadSHA:    req.PullRequest.HeadSHA,
		},
	}

	// Execute the use case
	result, err := h.planHandler.Handle(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert result to HTTP response
	response := PlanProjectResponse{
		Success:  result.Success,
		PlanFile: result.PlanFile,
		Output:   result.Output,
	}

	if result.Success {
		response.Message = "Project planned successfully"
	} else {
		response.Message = "Project planning failed"
	}

	w.Header().Set("Content-Type", "application/json")
	if result.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
	
	json.NewEncoder(w).Encode(response)
}

func (h *ProjectHandler) handleError(w http.ResponseWriter, err error) {
	response := PlanProjectResponse{
		Success: false,
		Message: err.Error(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
} 