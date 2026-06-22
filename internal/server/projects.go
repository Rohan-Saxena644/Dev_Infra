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
		http.Error(w,"invalid json", http.StatusBadRequest)
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
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}



func (s *Server) GetProjects(
	w http.ResponseWriter,
	r *http.Request,
){
	projects,err := s.DB.GetProjects(context.Background())
	if err != nil{
		http.Error(
			w,
			"failed to fetch projects",
			http.StatusInternalServerError,
		)

		return 
	}

	w.Header().Set("Content-type","application/json")
	json.NewEncoder(w).Encode(projects)
}




func (s *Server) GetProject(
	w http.ResponseWriter,
	r *http.Request,
)