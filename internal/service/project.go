package service

import (
	"context"

	"errors"
	"net/url"
	"strings"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProjectService struct {
	DB *database.Queries
}

func isValidGitHubRepoURL(repoURL string) bool {
	parsedURL, err := url.ParseRequestURI(repoURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "https" || parsedURL.Host != "github.com" || parsedURL.User != nil {
		return false
	}

	if parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return false
	}

	path := strings.TrimSuffix(strings.Trim(parsedURL.Path, "/"), ".git")
	parts := strings.Split(path, "/")

	return len(parts) == 2 &&
		parts[0] != "" && parts[0] != "." && parts[0] != ".." &&
		parts[1] != "" && parts[1] != "." && parts[1] != ".."
}

func (s *ProjectService) CreateProject(
	name string,
	repoURL string,
	userID int32,
) (database.Project, error) {
	if !isValidGitHubRepoURL(repoURL) {
		return database.Project{}, errors.New("invalid repository url")
	}

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

	count, err := s.DB.CountActiveDeploymentsByUser(
		context.Background(),
		pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	)

	if err != nil {
		return database.Deployment{}, err
	}

	if count >= 3 {
		return database.Deployment{}, errors.New("deployment limit reached")
	}

	_,err = s.DB.GetActiveDeploymentByProject(context.Background(),projectID)
	if err == nil {
		return database.Deployment{}, errors.New("deployment already running")
	}

	if err != pgx.ErrNoRows{
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
