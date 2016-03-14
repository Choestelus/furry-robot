package corgis

import (
	"bytes"
	"io"
	"os/exec"
	"sync"
	"time"
)

type JobType int

const (
	Streaming JobType = iota
	Timed     JobType = iota
	LRead     JobType = iota
	LWrite    JobType = iota
)

type JobScheduler struct {
	Cmd           *exec.Cmd
	Type          JobType
	LType         JobType
	ExecPeriod    time.Duration
	Wg            *sync.WaitGroup
	OutBuf        bytes.Buffer
	ErrBuf        bytes.Buffer
	CmdStatus     bool
	ProcessStatus bool
	Muffled       bool
	f             func(io.Reader, interface{}) []interface{}
	Res           []interface{}
}

type LatencyBucket struct {
	pid     float64
	latency float64
}
