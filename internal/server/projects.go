package server

import (
	// "context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w,"invalid json", http.StatusBadRequest)
		return
	}

	project, err := s.ProjectService.CreateProject(req.Name,req.RepoURL)

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
	projects,err := s.ProjectService.GetProjects()
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
){
	idstr := chi.URLParam(r,"id")
	id,err := strconv.Atoi(idstr)
	if err != nil{
		http.Error(
			w,
			"failed to parsde the id from the params",
			http.StatusInternalServerError,
		)
		return
	}


	project , err :=s.ProjectService.GetProject(int32(id))

	if err != nil{
		http.Error(
			w,
			"failed to get the id from the database",
			http.StatusInternalServerError,
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

	deployments, err := s.ProjectService.GetDeploymentsByProject(int32(id))
	if err != nil {
		http.Error(w, "failed to look up deployments for project", http.StatusInternalServerError)
		return
	}

	for _, d := range deployments {
		containerName := fmt.Sprintf("deployment-%d", d.ID)

		if out, err := s.Worker.Docker.Remove(containerName); err != nil {
			log.Println("delete project: failed to remove container", containerName, string(out), err)
		}
		if out, err := s.Worker.Docker.RemoveImage(containerName); err != nil {
			log.Println("delete project: failed to remove image", containerName, string(out), err)
		}
	}

	if err := s.ProjectService.DeleteProject(int32(id)); err != nil {
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}