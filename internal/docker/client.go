package docker

import (
	"os/exec"
)

type Client struct{}

func (c *Client)DockerPS()([]byte,error){
	return exec.Command(
		"docker",
		"ps",
	).CombinedOutput()
}


func (c *Client) Build(tag string, path string)([]byte,error){
	return exec.Command(
		"docker",
		"build",
		"-t",
		tag,
		path,
	).CombinedOutput()
}

func (c *Client) Run(containerName string, image string)([]byte,error){
	return exec.Command(
		"docker",
		"run",
		"-d",
		"--rm",
		"--name",
		containerName,
		"-p",
		"8081:80",
		image,
	).CombinedOutput()
}


func (c *Client) BuildRun(imageName string,containerName string,path string,)error{
	
}
