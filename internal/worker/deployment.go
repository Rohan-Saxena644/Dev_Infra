package worker

import (
	"time"
	"context"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
)


type DeploymentWorker struct{
	DB *database.Queries
}

func (w *DeploymentWorker)ProcessDeployment(
	deployemntID int32,
){
	w.DB.UpdateDeploymentStatus(context.Background(),
		database.UpdateDeploymentStatusParams{
			ID: deployemntID,
			Status: "running",
		},
	)	

	time.Sleep(5 * time.Second)

	w.DB.UpdateDeploymentStatus(
		context.Background(),
		database.UpdateDeploymentStatusParams{
			ID: deployemntID,
			Status: "success",
		},
	)
}