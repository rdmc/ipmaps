package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os/exec"
	"sync"
)

const (
	ipcliCmd   = "ipcli"
	listLeases = `list lease ip "remote id" iprange %s %s`
)

// ipRange defines a range of IPv4 address
type ipRange struct {
	Start, End net.IP
}

// CMCPEMap - CM mac -> CPE ip addresses Mapping
type CMCPEMap map[string]string

type ipMac struct {
	ip  string
	mac string
}

func getCMCPELeases() (CMCPEMap, error) {

	var err error

	var wg sync.WaitGroup
	resChan := make(chan ipMac)

	for _, r := range cfg.NetworkLeases {
		wg.Add(1)
		rr := r
		go func() {
			defer wg.Done()
			err = ipcliExec(rr, resChan)
			if err != nil {
				//log.Print("Error: ", err)
			}
			//fmt.Printf("[%s-%s] ", rr.Start.String(), rr.End.String())
		}()
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()
	var m CMCPEMap
	m = make(map[string]string)

	for r := range resChan {
		if prev, ok := m[r.mac]; ok {
			//fmt.Printf("DUP Mac, mac:%s, prev: %s, ip:%s\n", r.mac, prev, r.ip)
			m[r.mac] = fmt.Sprintf("%s;%s", prev, r.ip)

		} else {
			m[r.mac] = r.ip
		}
	}

	return m, err
}

func ipcliExec(ipr ipRange, rc chan<- ipMac) error {
	if _, err := exec.LookPath(ipcliCmd); err != nil {
		return fmt.Errorf("ipcliExec: %q executable not found, %v", cfg.IPCliCmd, err)
	}

	commandLine := fmt.Sprintf(listLeases, ipr.Start, ipr.End)

	args := []string{"-S", cfg.IPCliCluster,
		"-N", cfg.IPCliUser,
		"-P", cfg.IPCliPass,
		"-OF", "CSV"}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ipcliCmd, args...)
	cmd.Dir = cfg.WorkDir
	cmd.Stdin = bytes.NewBufferString(commandLine)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ipcliExec: stdiutpipe() error: %v", err)
	}
	reader := csv.NewReader(stdout)
	reader.FieldsPerRecord = 2

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ipcliExec: exec start() error: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		for {
			rec, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err == nil && len(rec[0]) > 6 && len(rec[1]) == 12 {
				rc <- ipMac{ip: rec[0], mac: rec[1]}
			}
		}
		wg.Done()
	}()

	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("exec wait() error: %v", err)
	}

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("timed out")
	}

	return nil
}
