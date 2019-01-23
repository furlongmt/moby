package daemon

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
)

func (daemon *Daemon) StartIter(ctx context.Context, containerID string) (container.IterBody, error) {
	fmt.Println("Starting iter...\n")
	return container.IterBody{}, nil
}

func (daemon *Daemon) StopIter(ctx context.Context, containerID string) (container.IterBody, error) {
	fmt.Println("Stopping iter...\n")
	return container.IterBody{}, nil
}
