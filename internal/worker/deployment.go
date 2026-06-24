package worker

import (
	"time"
	"context"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"log"
)


type DeploymentWorker struct{
	DB *database.Queries
	Queue chan int32
}

func (w *DeploymentWorker)ProcessDeployment(
	deploymentID int32,
){
	w.DB.UpdateDeploymentStatus(context.Background(),
		database.UpdateDeploymentStatusParams{
			ID: deploymentID,
			Status: "running",
		},
	)	

	log.Printf(
		"processing deployment %d",
		deploymentID,
	)

	time.Sleep(5 * time.Second)

	log.Printf(
		"deployment %d finished",
		deploymentID,
	)

	w.DB.UpdateDeploymentStatus(
		context.Background(),
		database.UpdateDeploymentStatusParams{
			ID: deploymentID,
			Status: "success",
		},
	)
}


func (w *DeploymentWorker) Start(){
	for deploymentID := range w.Queue{
		w.ProcessDeployment(deploymentID)
	}
}