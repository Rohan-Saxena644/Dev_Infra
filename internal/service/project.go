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


func (s *ProjectService) GetDeployment(id int32) (database.Deployment, error) {
	return s.DB.GetDeployment(context.Background(), id)
}


func (s *ProjectService) GetDeploymentsByProject(projectID int32) ([]database.Deployment, error) {
	return s.DB.GetDeploymentsByProject(context.Background(), projectID)
}


// DeleteProject removes a project's deployment rows first, then the
// project itself, since deployments.project_id has a foreign key
// pointing at projects.id with no cascade configured. Docker cleanup
// (stopping/removing containers and images) is the caller's
// responsibility — this only touches the database.
func (s *ProjectService) DeleteProject(projectID int32) error {
	if err := s.DB.DeleteDeploymentsByProject(context.Background(), projectID); err != nil {
		return err
	}
	return s.DB.DeleteProject(context.Background(), projectID)
}