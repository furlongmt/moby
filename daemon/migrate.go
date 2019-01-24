package daemon

// MATT ADDED THIS FILE

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/checkpoint-restore/go-criu/phaul"
	"github.com/docker/docker/api/types/container"
	"github.com/pkg/errors"
)

// TODO: make sure it's not a unix socket...
func getTCPHostAddress(daemon *Daemon) (string, error) {

	for k, _ := range daemon.hosts {
		host, _, err := net.SplitHostPort(k)
		if err == nil {
			return host, nil
		}
	}

	return "", errors.New("No tcp host found in docker daemon!")
}

func (daemon *Daemon) CreatePageServer(ctx context.Context, containerID string) (container.CreatePageServerBody, error) {
	fmt.Println("Creating page server...")
	// TODO: hard-coded port number
	port := uint32(6245)
	wdir, err := ioutil.TempDir("", "ctrd-pageserver-workdir")
	if err != nil {
		fmt.Println("Failed to create tmp dir")
		return container.CreatePageServerBody{}, err
	}

	for k, v := range daemon.hosts {
		fmt.Println("k:", k, "v:", v)
	}

	host, err := getTCPHostAddress(daemon)
	if err != nil {
		fmt.Println("Failed to get TCP Host Address")
		return container.CreatePageServerBody{}, err
	}

	fmt.Println(host)

	// TODO: validate s.ListenIP during initialization
	server, err := phaul.MakePhaulServer(phaul.Config{
		//Addr: daemon.hosts[0], // not sure about this
		Addr: "141.212.110.172",
		Port: int32(port),
		Wdir: wdir,
	})
	if err != nil {
		fmt.Println("Failed to make phaul server")
		return container.CreatePageServerBody{}, err
	}

	if daemon.pageServers == nil {
		daemon.pageServers = make(map[string]*phaul.Server)
	}

	daemon.pageServers[containerID] = server

	return container.CreatePageServerBody{
		Port: port,
	}, nil
}

func (daemon *Daemon) StartIter(ctx context.Context, containerID string) (container.IterBody, error) {
	fmt.Println("Starting iter...")
	return container.IterBody{}, nil
}

func (daemon *Daemon) StopIter(ctx context.Context, containerID string) (container.IterBody, error) {
	fmt.Println("Stopping iter...")
	return container.IterBody{}, nil
}
