package client

// MATT ADDED THIS FILE

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types/container"
)

func (cli *Client) CreatePageServer(ctx context.Context, containerID string) (container.CreatePageServerBody, error) {
	var response container.CreatePageServerBody

	fmt.Println("Creating page server")

	serverResp, err := cli.get(ctx, "/containers/"+containerID+"/createpageserver", nil, nil)

	if err != nil {
		return response, err
	}

	err = json.NewDecoder(serverResp.body).Decode(&response)
	ensureReaderClosed(serverResp)

	return response, err
}

func (c *Client) StartIter(ctx context.Context, containerID string) (container.IterBody, error) {
	fmt.Println("Starting iter...")
	return container.IterBody{}, nil
}

func (c *Client) StopIter(ctx context.Context, containerID string) (container.IterBody, error) {
	fmt.Println("Stopping iter...")
	return container.IterBody{}, nil
}
