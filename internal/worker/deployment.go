package worker

import (
	"context"
	"fmt"
	"time"

	"log"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/Rohan-Saxena644/devinfra/internal/docker"
	"github.com/Rohan-Saxena644/devinfra/internal/git"
)


type DeploymentWorker struct{
	DB *database.Queries
	Queue chan int32
	Docker *docker.Client
	Git *git.Client
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

	imageName := fmt.Sprintf("deployment-%d",deploymentID)

	containerName := imageName

	log.Printf("processing deployment %d",deploymentID)

	err := w.Docker.Deploy(
		imageName,
		containerName,
		"./test-app",
	)

	if err != nil {
		// update deployment to failed

		log.Println(err)

		w.DB.UpdateDeploymentStatus(
			context.Background(),
			database.UpdateDeploymentStatusParams{
				ID: deploymentID,
				Status: "failed",
			},
		)

		return
	}
	
	// output, err := w.Docker.Build(
	// 	imageName,
	// 	"./test-app",
	// )

	// log.Println(string(output))

	// if err != nil {
	// 	return
	// }

    // containerName := imageName

	// output, err = w.Docker.Run(
	// 	containerName,
	// 	imageName,
	// )

	// log.Println(string(output))

	// if err != nil {
	// 	log.Println(err)
	// }

	// log.Println(string(output))

	// log.Printf(
	// 	"processing deployment %d",
	// 	deploymentID,
	// )

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