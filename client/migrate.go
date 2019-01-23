package client

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
)

func (c *Client) StartIter(ctx context.Context, containerID string) (container.IterBody, error) {
	fmt.Println("Starting iter...\n")
	return container.IterBody{}, nil
}

func (c *Client) StopIter(ctx context.Context, containerID string) (container.IterBody, error) {
	fmt.Println("Stopping iter...\n")
	return container.IterBody{}, nil
}
