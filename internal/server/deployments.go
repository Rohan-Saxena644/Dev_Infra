package server

import (
	// "context"
	"encoding/json"
	"net/http"
	"strconv"

	// "github.com/Rohan-Saxena644/devinfra/internal/database"

	// "github.com/Rohan-Saxena644/devinfra/internal/service" was imported in the server struct from there the db is working entirely
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

	deployment, err := s.ProjectService.CreateDeployment(
		int32(id),
	)

	if err != nil {
		http.Error(
			w,
			"failed to create deployment",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}


func (s *Server) GetDeployments(
	w http.ResponseWriter,
	r *http.Request,
){

	deployments,err := s.ProjectService.GetDeployments()
	if err != nil{
		http.Error(
			w,
			"failed to get deployments",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployments)
}