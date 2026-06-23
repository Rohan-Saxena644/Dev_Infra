package repository

import "github.com/Rohan-Saxena644/devinfra/internal/database"

type PostgresProjectRepository struct {
	DB *database.Queries
}