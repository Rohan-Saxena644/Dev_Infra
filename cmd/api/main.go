package main

import (
	"context"
	"log"
	"net/http"
	"log/slog"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/Rohan-Saxena644/devinfra/internal/middleware"
	"github.com/Rohan-Saxena644/devinfra/internal/server"
	"github.com/Rohan-Saxena644/devinfra/internal/service"
	"github.com/Rohan-Saxena644/devinfra/internal/worker"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
	"os"
	"os/signal"
	"time"
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

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	r := chi.NewRouter()

	r.Use(middleware.Logging)

	r.Post("/projects",srv.CreateProject)
	r.Get("/projects", srv.GetProjects)
	r.Get("/projects/{id}",srv.GetProject)
	r.Post("/projects/{id}/deploy", srv.CreateDeployment)
	r.Get("/deployments", srv.GetDeployments)

	for i:=0;i<3;i++ {
		go worker.Start()
	}

	httpserver := &http.Server{
		Addr: ":8080",
		Handler: r,
	}

	log.Println("listening on :8080")
	// log.Fatal(http.ListenAndServe(":8080",r))

	go func() {
		if err := httpserver.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()


	<-ctx.Done()

	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := httpserver.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
