package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/lyft/temporal/workflows"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
)

func NewServerCmd() *cobra.Command {
	c := &cobra.Command{
		Use:           "application-server",
		Short:         "Start the application server",
		Long:          `Start the atlantis temporal application worker`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := NewServer()

			if err := s.server.ListenAndServe(); err != nil {
				panic(err)
			}

			return nil
		},
	}

	return c
}

type Server struct {
	temporal client.Client
	server   *http.Server
}

func NewServer() *Server {
	temporal, err := client.NewClient(client.Options{})
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	httpServer := &http.Server{
		Addr:    ":9000",
		Handler: r,
	}

	s := &Server{
		temporal: temporal,
		server:   httpServer,
	}

	// Backend
	r.Handle("/api/deploy", http.HandlerFunc(s.ExecuteQueueDeployment)).Methods(http.MethodPost)
	r.Handle("/api/plan_review", http.HandlerFunc(s.PlanReview)).Methods(http.MethodPost)

	return s

}

type ExecuteQueueDeploymentRequest struct {
	Revision string
	Repo     workflows.Repo
	Branch   string
}

type ExecuteQueueDeploymentResponse struct {
	WorkflowID string
	RunID      string
}

type PlanReviewRequest struct {
	User       string
	Status     workflows.PlanReviewStatus
	WorkflowID string
	RunID      string
}

func (s *Server) PlanReview(w http.ResponseWriter, r *http.Request) {
	var in PlanReviewRequest

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := s.temporal.SignalWorkflow(
		r.Context(),
		in.WorkflowID,
		in.RunID,
		workflows.PlanReviewSignal,
		workflows.PlanReview{
			User:   in.User,
			Status: in.Status,
		},
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) ExecuteQueueDeployment(w http.ResponseWriter, r *http.Request) {
	var in ExecuteQueueDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	options := client.StartWorkflowOptions{TaskQueue: workflows.TaskQueue}

	workflowRequest := workflows.DeployRequest{
		Repo:   in.Repo,
		Branch: in.Branch,
	}

	var deployWorkflow *workflows.Deploy

	// starts our workflow while also signaling it to a deploy a specific commit
	workflow, err := s.temporal.SignalWithStartWorkflow(
		r.Context(),
		fmt.Sprintf("%s-%s", in.Repo.Name, in.Branch),
		workflows.NewCommitAddedSignal,
		in.Revision,
		options,
		deployWorkflow.Run,
		workflowRequest,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&ExecuteQueueDeploymentResponse{
		WorkflowID: workflow.GetID(),
		RunID:      workflow.GetRunID(),
	})
}
