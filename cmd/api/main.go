package main

import (

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"github.com/Rohan-Saxena644/devinfra/internal/server"
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

	
}
