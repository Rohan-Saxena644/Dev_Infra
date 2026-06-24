package server

import (
	"github.com/Rohan-Saxena644/devinfra/internal/service"
	"github.com/Rohan-Saxena644/devinfra/internal/worker"
)

type Server struct{
	ProjectService *service.ProjectService
	Worker *worker.DeploymentWorker
}