package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Rohan-Saxena644/devinfra/internal/secrets"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
)

type EnvironmentVariableRequest struct {
	Value string `json:"value"`
}

type EnvironmentKeysResponse struct {
	Keys []string `json:"keys"`
}

func (s *Server) SetProjectEnvironmentVariable(w http.ResponseWriter, r *http.Request) {
	projectID, ok := projectEnvironmentID(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 20<<10)
	var request EnvironmentVariableRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	err := s.ProjectService.SetProjectEnvVar(projectID, chi.URLParam(r, "name"), request.Value)
	if writeEnvironmentError(w, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetProjectEnvironment(w http.ResponseWriter, r *http.Request) {
	projectID, ok := projectEnvironmentID(w, r)
	if !ok {
		return
	}

	keys, err := s.ProjectService.ListProjectEnvVarKeys(projectID)
	if writeEnvironmentError(w, err) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(EnvironmentKeysResponse{Keys: keys})
}

func (s *Server) DeleteProjectEnvironmentVariable(w http.ResponseWriter, r *http.Request) {
	projectID, ok := projectEnvironmentID(w, r)
	if !ok {
		return
	}

	err := s.ProjectService.DeleteProjectEnvVar(projectID, chi.URLParam(r, "name"))
	if writeEnvironmentError(w, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func projectEnvironmentID(w http.ResponseWriter, r *http.Request) (int32, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return 0, false
	}
	return int32(id), true
}

func writeEnvironmentError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, secrets.ErrInvalidVariable) {
		http.Error(w, "invalid environment variable", http.StatusBadRequest)
		return true
	}
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "project not found", http.StatusNotFound)
		return true
	}
	slog.Error("project environment operation failed", "error", err)
	http.Error(w, "environment operation failed", http.StatusInternalServerError)
	return true
}
