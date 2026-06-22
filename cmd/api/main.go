package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/Rohan-Saxena644/devinfra/internal/server"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
)

func main(){

	conn, err := pgx.Connect(context.Background(),"postgres://admin:admin@localhost:5433/devinfra?sslmode=disable")
	if err != nil{
		log.Fatal(err)
	}

	defer conn.Close(context.Background())
	queries := database.New(conn)

	srv := &server.Server{
		DB: queries,
	}

	r := chi.NewRouter()

	r.Post("/projects",srv.CreateProject)
	r.Get("/projects", srv.GetProjects)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080",r))
}
