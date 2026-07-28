package service

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"

	"github.com/Rohan-Saxena644/devinfra/internal/cache"
	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ProjectService struct {
	DB     *database.Queries
	Cache  *cache.Client
	EnvKey []byte
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

func (s *ProjectService) CreateProject(name string, repoURL string) (database.Project, error) {
	if !isValidGitHubRepoURL(repoURL) {
		return database.Project{}, errors.New("invalid repository url")
	}

	project, err := s.DB.CreateProject(context.Background(), database.CreateProjectParams{
		Name:    name,
		RepoUrl: repoURL,
	})
	if err == nil && s.Cache != nil {
		_ = s.Cache.DeleteProjects(context.Background())
	}
	return project, err
}

func (s *ProjectService) GetProjects() ([]database.Project, error) {
	ctx := context.Background()
	if s.Cache != nil {
		projects, found, err := s.Cache.GetProjects(ctx)
		if err == nil && found {
			return projects, nil
		}
		if err != nil {
			slog.Warn("project cache read failed", "error", err)
		}
	}

	projects, err := s.DB.GetProjects(ctx)
	if err == nil && s.Cache != nil {
		if cacheErr := s.Cache.SetProjects(ctx, projects); cacheErr != nil {
			slog.Warn("project cache write failed", "error", cacheErr)
		}
	}
	return projects, err
}

func (s *ProjectService) GetProject(id int32) (database.Project, error) {
	return s.DB.GetProject(context.Background(), id)
}

func (s *ProjectService) CreateDeployment(projectID int32) (database.Deployment, error) {
	if _, err := s.GetProject(projectID); err != nil {
		return database.Deployment{}, err
	}

	_, err := s.DB.GetActiveDeploymentByProject(context.Background(), projectID)
	if err == nil {
		return database.Deployment{}, errors.New("deployment already running")
	}
	if err != pgx.ErrNoRows {
		return database.Deployment{}, err
	}

	deployment, err := s.DB.CreateDeployment(context.Background(), database.CreateDeploymentParams{
		ProjectID: projectID,
		Status:    "queued",
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) &&
		(pgErr.ConstraintName == "deployments_global_limit" ||
			pgErr.ConstraintName == "deployments_daily_limit") {
		return database.Deployment{}, errors.New("deployment limit reached")
	}
	return deployment, err
}

func (s *ProjectService) GetDeployments() ([]database.Deployment, error) {
	return s.DB.GetDeployments(context.Background())
}

func (s *ProjectService) GetDeployment(id int32) (database.Deployment, error) {
	return s.DB.GetDeployment(context.Background(), id)
}

func (s *ProjectService) GetDeploymentsByProject(projectID int32) ([]database.Deployment, error) {
	if _, err := s.GetProject(projectID); err != nil {
		return nil, err
	}
	return s.DB.GetDeploymentsByProject(context.Background(), projectID)
}

// DeleteProject removes deployment resources in the handler before deleting
// their rows and the shared demo project.
func (s *ProjectService) DeleteProject(projectID int32) error {
	if _, err := s.GetProject(projectID); err != nil {
		return err
	}

	if err := s.DB.DeleteDeploymentsByProject(context.Background(), projectID); err != nil {
		return err
	}

	err := s.DB.DeleteProject(context.Background(), projectID)
	if err == nil && s.Cache != nil {
		_ = s.Cache.DeleteProjects(context.Background())
	}
	return err
}
