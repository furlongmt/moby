package daemon

// MATT ADDED THIS FILE

import (
	"context"
	"fmt"
	"net"
	"os"

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

func (daemon *Daemon) CreatePageServer(ctx context.Context, containerID string, wdir string) (container.CreatePageServerBody, error) {
	fmt.Println("Creating page server...")
	// TODO: hard-coded port number
	port := uint32(6245)
	// ignore errors cause either way we're making the dir
	_ = os.Mkdir(wdir, os.ModeDir)

	addr, err := getTCPHostAddress(daemon)
	if err != nil {
		logrus.Error("Failed to get TCP host address")
		return container.CreatePageServerBody{}, err
	}

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

	_, ok := daemon.pageServers[containerID]

	if ok { // if we previously had a page server for this container id
		logrus.Debugf("Deleting %v from map and killing page server", containerID)
		err = daemon.pageServers[containerID].KillPageServer()
		if err != nil {
			logrus.Error("Failed to kill previous page server!")
			return container.CreatePageServerBody{}, err
		}
		delete(daemon.pageServers, containerID)
	}

	daemon.pageServers[containerID] = server

	logrus.Debugf("Page server created on addr " + addr)

	return container.CreatePageServerBody{
		Port: port,
	}, nil
}

func (daemon *Daemon) StartIter(ctx context.Context, containerID string) error {
	logrus.Debugf("Starting iter...")
	err := daemon.pageServers[containerID].StartIter()
	if err != nil {
		logrus.Error("Failed to start page server iter")
		return err
	}

	return nil
}

func (daemon *Daemon) StopIter(ctx context.Context, containerID string) error {
	logrus.Debugf("Stopping iter...")
	err := daemon.pageServers[containerID].StopIter()
	if err != nil {
		logrus.Error("Failed to stop page server iter")
		return err
	}

	return nil
}

func (daemon *Daemon) MergeImages(ctx context.Context, containerId, dumpDir, lastDumpDir string) error {
	logrus.Debugf("Merging last dump dir %s into dump dir %s\n", lastDumpDir, dumpDir)
	return nil
}
