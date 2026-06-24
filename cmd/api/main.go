package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/Rohan-Saxena644/devinfra/internal/service"
	"github.com/Rohan-Saxena644/devinfra/internal/server"
	"github.com/Rohan-Saxena644/devinfra/internal/worker"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
)

func main(){

	conn, err := pgx.Connect(context.Background(),"postgres://admin:admin@localhost:15432/devinfra?sslmode=disable")
	if err != nil{
		log.Fatal(err)
	}

	defer conn.Close(context.Background())
	queries := database.New(conn)

	// srv := &server.Server{
	// 	DB: queries,
	// }

	projectService := &service.ProjectService{
		DB: queries,
	}

	worker := &worker.DeploymentWorker{
		DB: queries,
		Queue: make(chan int32, 100),
	}

	srv := &server.Server{
		ProjectService: projectService,
		Worker: worker,
	}

	r := chi.NewRouter()

	r.Post("/projects",srv.CreateProject)
	r.Get("/projects", srv.GetProjects)
	r.Get("/projects/{id}",srv.GetProject)
	r.Post("/projects/{id}/deploy", srv.CreateDeployment)
	r.Get("/deployments", srv.GetDeployments)

	for range 3 {
		go worker.Start()
	}
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080",r))
}
