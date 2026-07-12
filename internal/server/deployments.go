package server

import (
	// "context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/Rohan-Saxena644/devinfra/internal/middleware"

	// "github.com/Rohan-Saxena644/devinfra/internal/service" was imported in the server struct from there the db is working entirely
	//"github.com/Rohan-Saxena644/devinfra/internal/worker" same as above in the server struct it was imported the db is connected to the reference struct pointer
	"github.com/go-chi/chi"
)

func (s *Server) CreateDeployment(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(
			w,
			"invalid project id",
			http.StatusBadRequest,
		)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deployment, err := s.ProjectService.CreateDeployment(
		int32(id),
		userID,
	)

	// if err != nil {
	// 	slog.Error("failed to create deployment", "project_id", id, "error", err)
	// 	http.Error(
	// 		w,
	// 		"project not found",
	// 		http.StatusNotFound,
	// 	)
	// 	return
	// }

	if err != nil {

		if err.Error() == "deployment limit reached" {
			http.Error(
				w,
				"deployment limit reached",
				http.StatusTooManyRequests,
			)
			return
		}

		if err.Error() == "deployment already running" {
			http.Error(w, "deployment already running", http.StatusConflict)
			return
		}

		slog.Error("failed to create deployment", "project_id", id, "error", err)
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	s.Worker.Queue <- deployment.ID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}

type DeploymentResponse struct {
	database.Deployment
	ContainerRunning bool `json:"ContainerRunning"`
}

func (s *Server) GetDeployments(
	w http.ResponseWriter,
	r *http.Request,
) {

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deployments, err := s.ProjectService.GetDeployments(userID)
	if err != nil {
		slog.Error("failed to get deployments", "error", err)
		http.Error(
			w,
			"failed to get deployments",
			http.StatusInternalServerError,
		)
		return
	}

	response := make([]DeploymentResponse, 0, len(deployments))
	for _, d := range deployments {
		running := false
		if d.Status == "success" {
			containerName := fmt.Sprintf("deployment-%d", d.ID)
			running, _ = s.Worker.Docker.IsRunning(containerName)
		}
		response = append(response, DeploymentResponse{
			Deployment:       d,
			ContainerRunning: running,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) RestartDeployment(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid deployment id", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deployment, err := s.ProjectService.GetDeployment(int32(id), userID)
	if err != nil {
		slog.Error("deployment not found", "deployment_id", id, "error", err)
		http.Error(w, "deployment not found", http.StatusNotFound)
		return
	}

	if deployment.Status != "success" {
		http.Error(w, "only successful deployments can be restarted", http.StatusBadRequest)
		return
	}

	containerName := fmt.Sprintf("deployment-%d", deployment.ID)

	output, err := s.Worker.Docker.Start(containerName)
	if err != nil {
		log.Println("restart failed:", string(output), err)
		http.Error(w, "failed to restart container — it may have been permanently removed, redeploy instead", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DeploymentResponse{
		Deployment:       deployment,
		ContainerRunning: true,
	})
}
