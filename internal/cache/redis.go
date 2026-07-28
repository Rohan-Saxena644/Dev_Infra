package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	redis *redis.Client
}

const projectsKey = "projects:demo"

func New(addr string) *Client {
	return &Client{redis: redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		PoolSize:     5,
	})}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.redis.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.redis.Close()
}

func (c *Client) GetProjects(
	ctx context.Context,
) ([]database.Project, bool, error) {
	value, err := c.redis.Get(ctx, projectsKey).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var projects []database.Project
	if err := json.Unmarshal(value, &projects); err != nil {
		return nil, false, err
	}
	return projects, true, nil
}

func (c *Client) SetProjects(
	ctx context.Context,
	projects []database.Project,
) error {
	value, err := json.Marshal(projects)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, projectsKey, value, time.Minute).Err()
}

func (c *Client) DeleteProjects(ctx context.Context) error {
	return c.redis.Del(ctx, projectsKey).Err()
}
