package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Client struct{}

func (c *Client)DockerPS()([]byte,error){
	return exec.Command(
		"docker",
		"ps",
	).CombinedOutput()
}


func (c *Client) Build(tag string, path string)([]byte,error){

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()


	return exec.CommandContext(ctx,
		"docker",
		"build",
		"-t",
		tag,
		path,
	).CombinedOutput()
}


func (c *Client) Run(containerName string, image string , port int, containerPort int, envFile string)([]byte,error){

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	args := []string{
		"run",
		"-d",
		"--restart",
		"unless-stopped",
		"--memory",
		"512m",
		"--cpus",
		"0.50",
		"--pids-limit",
		"100",
	}
	if envFile != "" {
		args = append(args, "--env-file", envFile)
	}
	args = append(args,
		"--name",
		containerName,
		"-p",
		fmt.Sprintf("%d:%d", port, containerPort),
		image,
	)

	return exec.CommandContext(ctx, "docker", args...).CombinedOutput()
}


func (c *Client) Deploy(imageName string,containerName string,path string, port int, envFile string)error{
	_,err:= c.Build(imageName,path)
	if err != nil{
		return err
	}

	containerPort, err := c.ExposedPort(imageName)
	if err != nil {
		return err
	}

	_,err = c.Run(containerName,imageName, port, containerPort, envFile)
	if err != nil{
		return err
	}
	
	return nil
}

func (c *Client) ExposedPort(image string) (int, error) {
	output, err := exec.Command(
		"docker",
		"image",
		"inspect",
		"--format",
		"{{json .Config.ExposedPorts}}",
		image,
	).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("inspect image ports: %s: %w", strings.TrimSpace(string(output)), err)
	}

	var ports map[string]json.RawMessage
	if err := json.Unmarshal(output, &ports); err != nil {
		return 0, err
	}
	return selectExposedPort(ports)
}

func selectExposedPort(ports map[string]json.RawMessage) (int, error) {
	if _, ok := ports["80/tcp"]; ok {
		return 80, nil
	}

	var candidates []int
	for value := range ports {
		parts := strings.SplitN(value, "/", 2)
		if len(parts) != 2 || parts[1] != "tcp" {
			continue
		}
		port, err := strconv.Atoi(parts[0])
		if err == nil && port > 0 && port <= 65535 {
			candidates = append(candidates, port)
		}
	}
	if len(candidates) == 0 {
		return 80, nil
	}
	sort.Ints(candidates)
	if len(candidates) > 1 {
		return 0, errors.New("image exposes multiple TCP ports and none is port 80")
	}
	return candidates[0], nil
}


func (c *Client) Stop(containerName string) ([]byte, error) {
	return exec.Command(
		"docker",
		"stop",
		containerName,
	).CombinedOutput()
}

func (c *Client) Logs(containerName string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return exec.CommandContext(
		ctx,
		"docker",
		"logs",
		"--tail",
		"200",
		containerName,
	).CombinedOutput()
}


func (c *Client) Start(containerName string) ([]byte, error) {
	return exec.Command(
		"docker",
		"start",
		containerName,
	).CombinedOutput()
}


func (c *Client) Remove(containerName string) ([]byte, error) {
	return exec.Command(
		"docker",
		"rm",
		"-f",
		containerName,
	).CombinedOutput()
}


func (c *Client) RemoveImage(imageName string) ([]byte, error) {
	return exec.Command(
		"docker",
		"rmi",
		"-f",
		imageName,
	).CombinedOutput()
}


func (c *Client) IsRunning(containerName string) (bool, error) {
	out, err := exec.Command(
		"docker",
		"inspect",
		"-f",
		"{{.State.Running}}",
		containerName,
	).CombinedOutput()

	if err != nil {
		return false, nil
	}

	return strings.TrimSpace(string(out)) == "true", nil
}
