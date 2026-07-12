package git

import (
"context"
"time"
"os/exec"
)

type Client struct{}

func (c *Client) Clone(repoUrl, destination string)([]byte,error) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return exec.CommandContext(ctx,
		"git",
		"clone",
		repoUrl,
		destination,
	).CombinedOutput()

}

