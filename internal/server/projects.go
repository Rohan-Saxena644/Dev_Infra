package server

import (
	// "context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Rohan-Saxena644/devinfra/internal/middleware"

	// "github.com/Rohan-Saxena644/devinfra/internal/database"

	// "github.com/Rohan-Saxena644/devinfra/internal/service" was imported in the server struct from there the db is working entirely
	"github.com/go-chi/chi"
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

	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	project, err := s.ProjectService.CreateProject(req.Name, req.RepoURL, userID)

	if err != nil {
		if err.Error() == "invalid repository url" {
			http.Error(w, "invalid github repository url", http.StatusBadRequest)
			return
		}

		slog.Error("failed to create project", "error", err)
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
) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	projects, err := s.ProjectService.GetProjects(userID)
	if err != nil {
		slog.Error("failed to fetch projects", "error", err)
		http.Error(
			w,
			"failed to fetch projects",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (s *Server) GetProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	idstr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(
			w,
			"failed to parsde the id from the params",
			http.StatusInternalServerError,
		)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	project, err := s.ProjectService.GetProject(int32(id), userID)

	if err != nil {
		slog.Error("failed to get project", "id", id, "error", err)
		http.Error(
			w,
			"project not found",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)

}

func (s *Server) DeleteProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deployments, err := s.ProjectService.GetDeploymentsByProject(int32(id), userID)
	if err != nil {
		slog.Error("failed to look up deployments for project", "project_id", id, "error", err)
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	for _, d := range deployments {
		if out, err := s.Worker.RemoveDeployment(d); err != nil {
			log.Println("delete project: failed to remove deployment", d.ID, string(out), err)
		}
	}

	if err := s.ProjectService.DeleteProject(int32(id), userID); err != nil {
		slog.Error("failed to delete project", "project_id", id, "error", err)
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
