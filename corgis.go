package corgis

import (
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type JobType int

const (
	Streaming JobType = iota
	Timed     JobType = iota
)

func timedSIGTERM(p *os.Process, d time.Duration) {
	var _ = <-time.After(d)
	err := p.Signal(syscall.SIGTERM)
	if err != nil {
		log.Panicf("timedSIGTERM Panicking: %v\n", err)
	}
}

type JobScheduler struct {
	Cmd           *exec.Cmd
	Type          JobType
	ExecPeriod    time.Duration
	Wg            *sync.WaitGroup
	OutPipe       *io.ReadCloser
	ErrPipe       *io.ReadCloser
	CmdStatus     bool
	ProcessStatus bool
}

func (j *JobScheduler) renewCmd() {
	j.Cmd = exec.Command(j.Cmd.Path, j.Cmd.Args[1])
	outp, err := j.Cmd.StdoutPipe()
	if err != nil {
		log.Panicf("renewCmd set stdout error: %v\n", err)
	}
	j.OutPipe = &outp
	oute, err := j.Cmd.StderrPipe()
	if err != nil {
		log.Panicf("renewCmd set stderr error: %v\n", err)
	}
	j.ErrPipe = &oute
}
