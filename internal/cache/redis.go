package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	redis *redis.Client
}

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
	userID int32,
) ([]database.Project, bool, error) {
	value, err := c.redis.Get(ctx, projectsKey(userID)).Bytes()
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
	userID int32,
	projects []database.Project,
) error {
	value, err := json.Marshal(projects)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, projectsKey(userID), value, time.Minute).Err()
}

func (c *Client) DeleteProjects(ctx context.Context, userID int32) error {
	return c.redis.Del(ctx, projectsKey(userID)).Err()
}

func projectsKey(userID int32) string {
	return fmt.Sprintf("projects:user:%d", userID)
}
