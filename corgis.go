package corgis

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type JobType int

const (
	Streaming JobType = iota
	Timed     JobType = iota
)

func timedSIGTERM(p *os.Process, d time.Duration) {
	log.Printf("timedSIGTERM called\n")
	var _ = <-time.After(d)
	err := p.Signal(syscall.SIGTERM)
	log.Printf("SIGTERM sent\n")
	if err != nil {
		log.Panicf("timedSIGTERM Panicking: %v\n", err)
	}
}

func TimedSIGTERM(p *os.Process, d time.Duration) {
	timedSIGTERM(p, d)
}

func DecodeStream(r io.Reader) []interface{} {
	decoder := json.NewDecoder(r)
	openToken, err := decoder.Token()
	if err != nil {
		log.Fatalf("openToken error: %v\n", openToken)
	}
	//fmt.Println("openToken:", openToken)
	res := make([]interface{}, 0)
	for decoder.More() {
		var m interface{}
		err := decoder.Decode(&m)
		if err != nil {
			log.Fatalf("error decoding: %v\n", err)
		}
		res = append(res, m)
		fmt.Printf("decoded: [%v]\n", m)
	}
	closeToken, err := decoder.Token()
	if err != nil {
		log.Fatalf("closeToken error: %v\n", closeToken)
	}
	//fmt.Println("closeToken:", closeToken)
	return res
}

func (j *JobScheduler) InitCmd() {
	j.Cmd = exec.Command(j.Cmd.Path, j.Cmd.Args[1])
	j.OutBuf.Reset()
	j.ErrBuf.Reset()
	j.Cmd.Stdout = &j.OutBuf
	j.Cmd.Stderr = &j.ErrBuf
	log.Printf("Initialized job\n")
}

func (j *JobScheduler) ExecStreaming() {

}

func (j *JobScheduler) ExecTimed() {
	err := j.Cmd.Start()
	if err != nil {
		log.Panicf("cmd start error: %v\n", err)
	}
	log.Printf("job started\n")
	go timedSIGTERM(j.Cmd.Process, j.ExecPeriod)
	log.Printf("waiting for job terminate\n")
	err = j.Cmd.Wait()
	if err != nil {
		log.Panicf("cmd wait error: %v\n", err)
	}
	if j.Cmd.ProcessState.Success() {
		j.Res = DecodeStream(&j.OutBuf)
		j.InitCmd()
		j.ExecTimed()
	}
}

func (j *JobScheduler) Execute() {
	if j.Type == Streaming {
		j.ExecStreaming()

	} else if j.Type == Timed {
		j.ExecTimed()
	}
}
