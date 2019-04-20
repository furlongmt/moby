package phaul

import (
	"fmt"

	criu "github.com/checkpoint-restore/go-criu"
	"github.com/checkpoint-restore/go-criu/rpc"
	"github.com/checkpoint-restore/go-criu/stats"
	"github.com/golang/protobuf/proto"
)

const minPagesWritten uint64 = 64 // minimum number of dumped pages to continue iteration
const maxIters int = 8            // maximum number of iterations
const maxGrowDelta float64 = 10   // maximum acceptable growth rate in percentage

// Client struct
type Client struct {
	local  Local
	remote Remote
	cfg    Config
}

// MakePhaulClient function
// Main entry point. Caller should create the client object by
// passing here local, remote and comm. See comment in corresponding
// interfaces/structs for explanation.
//
// Then call client.Migrate() and enjoy :)
func MakePhaulClient(l Local, r Remote, c Config) (*Client, error) {
	return &Client{local: l, remote: r, cfg: c}, nil
}

func calcGrowRate(value int64, prevValue int64) float64 {
	delta := value - prevValue
	return float64(delta) * 100 / float64(prevValue)
}

func isLastIter(iter int, stats *stats.DumpStatsEntry, prevStats *stats.DumpStatsEntry) bool {

	if iter >= maxIters {
		fmt.Printf("`- max iters reached, iter (%d)\n", iter)
		return true
	}

	pagesWritten := stats.GetPagesWritten()
	if pagesWritten < minPagesWritten {
		fmt.Printf("`- tiny pre-dump (%d) reached, iter (%d)\n", iter, int(pagesWritten))
		return true
	}

	prevPagesWritten := prevStats.GetPagesWritten()
	if prevPagesWritten != 0 { // not empty

		pagesDelta := calcGrowRate(int64(pagesWritten), int64(prevPagesWritten))
		if pagesDelta > maxGrowDelta {
			fmt.Printf("`- grow iter (%d) reached, iter(%d)\n", iter, int(pagesDelta))
			return true
		}
	}

	return false
}

// Migrate function
func (pc *Client) Migrate() error {
	criu := criu.MakeCriu()
	prevP := ""

	psi := rpc.CriuPageServerInfo{
		Address: proto.String(pc.cfg.Addr),
		Port:    proto.Int32(int32(pc.cfg.Port)),
	}
	opts := rpc.CriuOpts{
		Pid:      proto.Int32(int32(pc.cfg.Pid)),
		LogLevel: proto.Int32(4),
		LogFile:  proto.String("pre-dump.log"),
		Ps:       &psi,
	}

	err := criu.Prepare()
	if err != nil {
		return err
	}

	defer criu.Cleanup()

	imgs, err := preparePhaulImages(pc.cfg.Wdir)
	if err != nil {
		return err
	}
	prevStats := &stats.DumpStatsEntry{}
	iter := 0

	if pc.cfg.PreDump {
		for {
			err = pc.remote.StartIter()
			if err != nil {
				return err
			}

			prevP = imgs.lastImagesDir()
			imgDir, err := imgs.openNextDir()
			if err != nil {
				return err
			}

			opts.ImagesDirFd = proto.Int32(int32(imgDir.Fd()))
			if prevP != "" {
				opts.ParentImg = proto.String(prevP)
			}

			err = criu.PreDump(opts, nil)
			imgDir.Close()
			if err != nil {
				return err
			}

			err = pc.remote.StopIter()
			if err != nil {
				return err
			}

			st, err := criuGetDumpStats(imgDir)
			if err != nil {
				return err
			}

			if isLastIter(iter, st, prevStats) {
				break
			}

			prevStats = st
			iter++
		}
	}

	err = pc.remote.StartIter()
	if err == nil {
		prevP := imgs.lastImagesDir()
		err = pc.local.DumpCopyRestore(criu, pc.cfg, prevP)
		err2 := pc.remote.StopIter()
		if err == nil {
			err = err2
		}
	}

	return err
}
