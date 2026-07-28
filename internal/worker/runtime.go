package worker

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/Rohan-Saxena644/devinfra/internal/docker"
)

func deploymentName(deploymentID int32) string {
	return fmt.Sprintf("deployment-%d", deploymentID)
}

func (w *DeploymentWorker) IsDeploymentRunning(deployment database.Deployment) (bool, error) {
	if deployment.DeploymentType == "compose" {
		return w.Docker.ComposeIsRunning(
			docker.ComposeConfigPath(deployment.ID),
			docker.ComposeProjectName(deployment.ID),
		)
	}
	return w.Docker.IsRunning(deploymentName(deployment.ID))
}

func (w *DeploymentWorker) StopDeployment(deployment database.Deployment) ([]byte, error) {
	if deployment.DeploymentType == "compose" {
		return w.Docker.ComposeStop(
			docker.ComposeConfigPath(deployment.ID),
			docker.ComposeProjectName(deployment.ID),
		)
	}
	return w.Docker.Stop(deploymentName(deployment.ID))
}

func (w *DeploymentWorker) StartDeployment(deployment database.Deployment) ([]byte, error) {
	if deployment.DeploymentType == "compose" {
		return w.Docker.ComposeStart(
			docker.ComposeConfigPath(deployment.ID),
			docker.ComposeProjectName(deployment.ID),
		)
	}
	return w.Docker.Start(deploymentName(deployment.ID))
}

func (w *DeploymentWorker) DeploymentLogs(deployment database.Deployment) ([]byte, error) {
	if deployment.DeploymentType == "compose" {
		return w.Docker.ComposeLogs(
			docker.ComposeConfigPath(deployment.ID),
			docker.ComposeProjectName(deployment.ID),
		)
	}
	return w.Docker.Logs(deploymentName(deployment.ID))
}

func (w *DeploymentWorker) RemoveDeployment(deployment database.Deployment) ([]byte, error) {
	if deployment.DeploymentType == "compose" {
		configPath := docker.ComposeConfigPath(deployment.ID)
		if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		output, err := w.Docker.ComposeRemove(configPath, docker.ComposeProjectName(deployment.ID))
		if err == nil {
			_ = os.Remove(configPath)
		}
		return output, err
	}

	name := deploymentName(deployment.ID)
	containerOutput, containerErr := w.Docker.Remove(name)
	imageOutput, imageErr := w.Docker.RemoveImage(name)
	if strings.Contains(string(containerOutput), "No such container") {
		containerErr = nil
	}
	if strings.Contains(string(imageOutput), "No such image") {
		imageErr = nil
	}
	return bytes.Join([][]byte{containerOutput, imageOutput}, []byte("\n")), errors.Join(containerErr, imageErr)
}
