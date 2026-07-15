package service

import (
	"context"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/Rohan-Saxena644/devinfra/internal/secrets"
)

func (s *ProjectService) SetProjectEnvVar(
	projectID int32,
	userID int32,
	name string,
	value string,
) error {
	if _, err := s.GetProject(projectID, userID); err != nil {
		return err
	}
	if err := secrets.ValidateVariable(name, value); err != nil {
		return err
	}

	encrypted, err := secrets.Encrypt(s.EnvKey, value)
	if err != nil {
		return err
	}
	return s.DB.UpsertProjectEnvVar(context.Background(), database.UpsertProjectEnvVarParams{
		ProjectID:      projectID,
		Name:           name,
		EncryptedValue: encrypted,
	})
}

func (s *ProjectService) ListProjectEnvVarKeys(projectID int32, userID int32) ([]string, error) {
	if _, err := s.GetProject(projectID, userID); err != nil {
		return nil, err
	}

	variables, err := s.DB.ListProjectEnvVars(context.Background(), projectID)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(variables))
	for _, variable := range variables {
		keys = append(keys, variable.Name)
	}
	return keys, nil
}

func (s *ProjectService) DeleteProjectEnvVar(projectID int32, userID int32, name string) error {
	if _, err := s.GetProject(projectID, userID); err != nil {
		return err
	}
	if err := secrets.ValidateVariable(name, ""); err != nil {
		return err
	}
	return s.DB.DeleteProjectEnvVar(context.Background(), database.DeleteProjectEnvVarParams{
		ProjectID: projectID,
		Name:      name,
	})
}
