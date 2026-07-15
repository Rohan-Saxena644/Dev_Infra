package worker

import (
	"context"
	"fmt"
	"time"

	"log"
	"os"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/Rohan-Saxena644/devinfra/internal/docker"
	"github.com/Rohan-Saxena644/devinfra/internal/git"
	"github.com/jackc/pgx/v5/pgtype"
)


type DeploymentWorker struct{
	DB *database.Queries
	Queue chan int32
	Docker *docker.Client
	Git *git.Client
	EnvKey []byte
}



func (w *DeploymentWorker) markFailed(deploymentID int32) {
	_, err := w.DB.UpdateDeploymentStatus(
		context.Background(),
		database.UpdateDeploymentStatusParams{
			ID:     deploymentID,
			Status: "failed",
		},
	)
	if err != nil {
		log.Println("failed to mark deployment as failed:", err)
	}
}



func (w *DeploymentWorker) enforceFifoLimit(projectID int32, limit int) {
	deployments, err := w.DB.GetDeploymentsByProject(context.Background(), projectID)
	if err != nil {
		log.Println("fifo: failed to list deployments for project", projectID, err)
		return
	}

	if len(deployments) <= limit {
		return
	}

	for _, d := range deployments[limit:] {
		if out, err := w.RemoveDeployment(d); err != nil {
			log.Println("fifo: failed to remove deployment", d.ID, string(out), err)
		}
		if err := w.DB.DeleteDeployment(context.Background(), d.ID); err != nil {
			log.Println("fifo: failed to delete deployment row", d.ID, err)
		}
	}
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


	deployment, err := w.DB.GetDeployment(context.Background(),deploymentID)
	if err != nil{
		log.Println(err)
		w.markFailed(deploymentID)
		return 
	}

	project, err := w.DB.GetProject(context.Background(),deployment.ProjectID)
	if err != nil{
		log.Println(err)
		w.markFailed(deploymentID)
		return
	}

	envFile, environment, err := w.environmentFile(project.ID, deploymentID)
	if err != nil {
		log.Println("failed to prepare deployment environment:", err)
		w.markFailed(deploymentID)
		return
	}
	if envFile != "" {
		defer os.Remove(envFile)
	}

	path := fmt.Sprintf("./tmp/deployment-%d",deploymentID)

	defer os.RemoveAll(path)

	port := 9000 + int(deploymentID)

	err = w.DB.UpdateDeploymentPort(
		context.Background(),
		database.UpdateDeploymentPortParams{
			ID: deploymentID,
			Port: pgtype.Int4{
				Int32: int32(port),
				Valid: true,
			},
		},
	)

	if err != nil{
		log.Println(err)
		w.markFailed(deploymentID)
		return 
	}


	output, err := w.Git.Clone(project.RepoUrl,path)

	if err != nil{
		log.Println(string(output))
		log.Println(err)
		w.markFailed(deploymentID)
		return
	}


	log.Println("Clone complete")
	log.Println("Calling Docker.Deploy")


	imageName := fmt.Sprintf("deployment-%d",deploymentID)

	containerName := imageName

	log.Printf("processing deployment %d",deploymentID)

	composeFile, isCompose := docker.FindComposeFile(path)
	if isCompose {
		err = w.DB.UpdateDeploymentType(
			context.Background(),
			database.UpdateDeploymentTypeParams{
				ID:             deploymentID,
				DeploymentType: "compose",
			},
		)
		if err == nil {
			err = w.Docker.DeployCompose(
				composeFile,
				path,
				docker.ComposeProjectName(deploymentID),
				docker.ComposeConfigPath(deploymentID),
				port,
				environment,
			)
		}
	} else {
		err = w.Docker.Deploy(
			imageName,
			containerName,
			path,
			port,
			envFile,
		)
	}

	if err != nil {
		// update deployment to failed
		log.Println(err)
		w.markFailed(deploymentID)
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


	log.Println("Updating status to success")

	_,err = w.DB.UpdateDeploymentStatus(
		context.Background(),
		database.UpdateDeploymentStatusParams{
			ID: deploymentID,
			Status: "success",
		},
	)

	if err != nil{
		log.Println(err)
		return
	}

	w.enforceFifoLimit(project.ID, 10)
}




func (w *DeploymentWorker) Start(){

	log.Println("worker goroutine started")

	for deploymentID := range w.Queue{
		w.ProcessDeployment(deploymentID)
	}
}
