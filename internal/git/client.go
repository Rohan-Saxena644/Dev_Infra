package git

import "os/exec"

type Client struct{}

func (c *Client) Clone(repoUrl, destination string)([]byte,error) {

	return exec.Command(
		"git",
		"clone",
		repoUrl,
		destination,
	).CombinedOutput()

}

