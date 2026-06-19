package handlers

import(
	"net/http"
	"github.com/Rohan-Saxena644/devinfra/internal/server"
)

type CreateProjectRequest struct{
	Name string `json:"name"`
	RepoUrl string `json:"repo_url"`
}


func (s *Server) CreateProject(w http.ResponseWriter,r *http.Request){
	
}