package service

import (
	"context"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProjectService struct {
	DB *database.Queries
}

func (s *ProjectService) CreateProject(
	name string,
	repoURL string,
	userID int32,
) (database.Project, error) {
	return s.DB.CreateProject(context.Background(), database.CreateProjectParams{
		Name:    name,
		RepoUrl: repoURL,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})
}

func (s *ProjectService) GetProjects(userID int32) ([]database.Project, error) {
	return s.DB.GetProjectsByUser(context.Background(), pgtype.Int4{
		Int32: userID,
		Valid: true,
	})
}

func (s *ProjectService) GetProject(id int32, userID int32) (database.Project, error) {
	return s.DB.GetProjectByUser(context.Background(), database.GetProjectByUserParams{
		ID: id,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})
}

func (s *ProjectService) CreateDeployment(projectID int32, userID int32) (database.Deployment, error) {
	_, err := s.GetProject(projectID, userID)
	if err != nil {
		return database.Deployment{}, err
	}

	return s.DB.CreateDeployment(context.Background(), database.CreateDeploymentParams{
		ProjectID: projectID,
		Status:    "queued",
	})
}

func (s *ProjectService) GetDeployments(userID int32) ([]database.Deployment, error) {
	return s.DB.GetDeploymentsByUser(context.Background(), pgtype.Int4{
		Int32: userID,
		Valid: true,
	})
}

func (s *ProjectService) GetDeployment(id int32, userID int32) (database.Deployment, error) {
	deployment, err := s.DB.GetDeployment(context.Background(), id)
	if err != nil {
		return database.Deployment{}, err
	}

	_, err = s.GetProject(deployment.ProjectID, userID)
	if err != nil {
		return database.Deployment{}, err
	}

	return deployment, nil
}

func (s *ProjectService) GetDeploymentsByProject(projectID int32, userID int32) ([]database.Deployment, error) {
	_, err := s.GetProject(projectID, userID)
	if err != nil {
		return nil, err
	}

	return s.DB.GetDeploymentsByProject(context.Background(), projectID)
}

// DeleteProject removes a project's deployment rows first, then the
// project itself, since deployments.project_id has a foreign key
// pointing at projects.id with no cascade configured. Docker cleanup
// (stopping/removing containers and images) is the caller's
// responsibility — this only touches the database.
func (s *ProjectService) DeleteProject(projectID int32, userID int32) error {
	_, err := s.GetProject(projectID, userID)
	if err != nil {
		return err
	}

	if err := s.DB.DeleteDeploymentsByProject(context.Background(), projectID); err != nil {
		return err
	}

	return s.DB.DeleteProjectByUser(context.Background(), database.DeleteProjectByUserParams{
		ID: projectID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})
}
