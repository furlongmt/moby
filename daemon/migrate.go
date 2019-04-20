package daemon

// MATT ADDED THIS FILE

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"syscall"

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
			//return container.CreatePageServerBody{}, err
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

// TODO: this should really just be a restore really
func (daemon *Daemon) MergeImages(ctx context.Context, containerId, dumpDir string) (container.MergeImagesBody, error) {
	// probably don't need last dump dir to be passed in here...
	logrus.Debugf("IN MERGE \n\n")
	lastPreDumpDir := daemon.pageServers[containerId].LastImagesDir()
	// TODO: hacky...
	dumpDir = "/" + dumpDir // since it's tmp
	logrus.Debugf("MergeImages: dumpDir - " + dumpDir + " lastDumpDir - " + lastPreDumpDir)
	idir, err := os.Open(dumpDir)
	if err != nil {
		return container.MergeImagesBody{}, err
	}

	defer idir.Close()

	imgs, err := idir.Readdirnames(0)
	if err != nil {
		return container.MergeImagesBody{}, err
	}

	for _, fname := range imgs {
		if !strings.HasSuffix(fname, ".img") {
			continue
		}

		fmt.Printf("\t%s -> %s/\n", fname, lastPreDumpDir)
		err = syscall.Link(dumpDir+"/"+fname, lastPreDumpDir+"/"+fname)
		if err != nil {
			return container.MergeImagesBody{}, err
		}
	}

	return container.MergeImagesBody{
		Dir: lastPreDumpDir,
	}, err
}
