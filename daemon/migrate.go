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
	"github.com/sirupsen/logrus"
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
		logrus.Error("Failed to crate tmp dir for page server")
		return container.CreatePageServerBody{}, err
	}

	addr, err := getTCPHostAddress(daemon)
	if err != nil {
		logrus.Error("Failed to get TCP host address")
		return container.CreatePageServerBody{}, err
	}

	fmt.Println("Page server created on addr " + addr)

	// TODO: validate s.ListenIP during initialization
	server, err := phaul.MakePhaulServer(phaul.Config{
		Addr: addr,
		Port: int32(port),
		Wdir: wdir,
	})
	if err != nil {
		logrus.Error("Failed to create phaul server")
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

func (daemon *Daemon) StartIter(ctx context.Context, containerID string) error {
	fmt.Println("Starting iter...")
	err := daemon.pageServers[containerID].StartIter()
	if err != nil {
		logrus.Error("Failed to start page server iter")
		return err
	}

	return nil
}

func (daemon *Daemon) StopIter(ctx context.Context, containerID string) error {
	fmt.Println("Stopping iter...")
	err := daemon.pageServers[containerID].StopIter()
	if err != nil {
		logrus.Error("Failed to stop page server iter")
		return err
	}

	return nil
}
