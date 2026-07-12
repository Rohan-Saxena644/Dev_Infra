package docker

import (
	"os/exec"
	"fmt"
	"strings"
	"context"
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


func (c *Client) Run(containerName string, image string , port int)([]byte,error){

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	return exec.CommandContext(ctx,
		"docker",
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
		"--name",
		containerName,
		"-p",
		fmt.Sprintf("%d:80", port),
		image,
	).CombinedOutput()
}


func (c *Client) Deploy(imageName string,containerName string,path string, port int)error{
	_,err:= c.Build(imageName,path)
	if err != nil{
		return err
	}

	_,err = c.Run(containerName,imageName, port)
	if err != nil{
		return err
	}
	
	return nil
}


func (c *Client) Stop(containerName string) ([]byte, error) {
	return exec.Command(
		"docker",
		"stop",
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
