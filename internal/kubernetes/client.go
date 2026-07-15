package kubernetes

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Namespace string
}

func (c *Client) run(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if c.Namespace != "" {
		args = append([]string{"--namespace", c.Namespace}, args...)
	}

	output, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("kubectl failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return output, nil
}

func (c *Client) Deploy(name string, image string, port int) error {
	if _, err := c.run("create", "deployment", name, "--image", image); err != nil {
		return err
	}
	_, err := c.run("expose", "deployment", name, "--type", "ClusterIP", "--port", strconv.Itoa(port))
	return err
}

func (c *Client) Scale(name string, replicas int) ([]byte, error) {
	return c.run("scale", "deployment", name, "--replicas", strconv.Itoa(replicas))
}

func (c *Client) Logs(name string) ([]byte, error) {
	return c.run("logs", "deployment/"+name, "--tail", "200")
}

func (c *Client) Delete(name string) ([]byte, error) {
	return c.run("delete", "deployment,service", name, "--ignore-not-found=true")
}
