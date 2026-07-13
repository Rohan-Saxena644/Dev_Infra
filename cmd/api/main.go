package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"os"
	"os/signal"
	"time"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/Rohan-Saxena644/devinfra/internal/docker"
	"github.com/Rohan-Saxena644/devinfra/internal/git"
	"github.com/Rohan-Saxena644/devinfra/internal/middleware"
	"github.com/Rohan-Saxena644/devinfra/internal/server"
	"github.com/Rohan-Saxena644/devinfra/internal/service"
	"github.com/Rohan-Saxena644/devinfra/internal/worker"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load(".env.local")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	dbpool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	queries := database.New(dbpool)

	// srv := &server.Server{
	// 	DB: queries,
	// }

	projectService := &service.ProjectService{
		DB: queries,
	}

	dockerClient := &docker.Client{}
	git := &git.Client{}

	worker := &worker.DeploymentWorker{
		DB:     queries,
		Queue:  make(chan int32, 100),
		Docker: dockerClient,
		Git:    git,
	}

	srv := &server.Server{
		ProjectService: projectService,
		Worker:         worker,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	r := chi.NewRouter()

	r.Use(middleware.Cors)
	r.Use(middleware.Logging)

	r.Post("/auth/signup", srv.SignUp)
	r.Post("/auth/login", srv.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)

		r.Post("/projects", srv.CreateProject)
		r.Get("/projects", srv.GetProjects)
		r.Get("/projects/{id}", srv.GetProject)
		r.Delete("/projects/{id}", srv.DeleteProject)
		r.Post("/projects/{id}/deploy", srv.CreateDeployment)
		r.Get("/deployments", srv.GetDeployments)
		r.Post("/deployments/{id}/restart", srv.RestartDeployment)
	})

	for i := 0; i < 3; i++ {
		go worker.Start()
	}

	httpserver := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	log.Println("listening on :8080")
	// log.Fatal(http.ListenAndServe(":8080",r))

	// Graceful shutdoen

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
