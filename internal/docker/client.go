package docker

import (
	"os/exec"
	"fmt"
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

func (c *Client) Run(containerName string, image string , port int)([]byte,error){
	return exec.Command(
		"docker",
		"run",
		"-d",
		"--rm",
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
