package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
)

type CreateProjectRequest struct {
	Name    string `json:"name"`
	RepoURL string `json:"repo_url"`
}

func (s *Server) CreateProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req CreateProjectRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	project, err := s.DB.CreateProject(
		context.Background(),
		database.CreateProjectParams{
			Name:    req.Name,
			RepoUrl: req.RepoURL,
		},
	)

	if err != nil {
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}