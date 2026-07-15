package worker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Rohan-Saxena644/devinfra/internal/secrets"
)

func (w *DeploymentWorker) environmentFile(projectID int32, deploymentID int32) (string, map[string]string, error) {
	rows, err := w.DB.ListProjectEnvVars(context.Background(), projectID)
	if err != nil || len(rows) == 0 {
		return "", nil, err
	}
	if len(rows) > 50 {
		return "", nil, errors.New("project has too many environment variables")
	}

	variables := make(map[string]string, len(rows))
	for _, row := range rows {
		value, err := secrets.Decrypt(w.EnvKey, row.EncryptedValue)
		if err != nil {
			return "", nil, fmt.Errorf("decrypt environment variable %s: %w", row.Name, err)
		}
		variables[row.Name] = value
	}

	path, err := filepath.Abs(filepath.Join("tmp", fmt.Sprintf("deployment-%d.env", deploymentID)))
	if err != nil {
		return "", nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", nil, err
	}
	if err := secrets.WriteEnvFile(path, variables); err != nil {
		return "", nil, err
	}
	return path, variables, nil
}
