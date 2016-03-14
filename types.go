package corgis

import (
	"bytes"
	"io"
	"os/exec"
	"sync"
	"time"
)

type JobScheduler struct {
	Cmd           *exec.Cmd
	Type          JobType
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
