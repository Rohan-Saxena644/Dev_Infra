package service

import (
	"context"
	"github.com/Rohan-Saxena644/devinfra/internal/database"
)

type ProjectService struct {
	DB *database.Queries
}

func (s *ProjectService) CreateProject(
	name string,
	repoURL string,
)(database.Project,error){
	return s.DB.CreateProject(context.Background(),database.CreateProjectParams{
		Name: name,
		RepoUrl: repoURL,
	})
}


func (s *ProjectService) GetProjects() ([]database.Project, error) {
	return s.DB.GetProjects(context.Background())
}



func (s *ProjectService) GetProject(id int32) (database.Project, error) {
	return s.DB.GetProject(context.Background(), id)
}


func (s *ProjectService) CreateDeployment(projectID int32)(database.Deployment,error){
	return s.DB.CreateDeployment(context.Background(),database.CreateDeploymentParams{
		ProjectID: projectID,
		Status: "queued",
	})
}


func (s *ProjectService) GetDeployments() (
	[]database.Deployment,
	error,
) {
	return s.DB.GetDeployments(context.Background())
}