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