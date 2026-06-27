-- +goose Up
ALTER TABLE deployments
ADD COLUMN port INT;

