package corgis

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gopkg.in/pipe.v2"
)

var (
	DB *gorm.DB
)

func init() {
	var err error
	DB, err = gorm.Open("postgres", "user=postgres password=12344321 dbname=tiramisu sslmode=disable")
	if err != nil {
		log.Fatalf("dbconn error %v\n", err)
	}
	DB.AutoMigrate(&RawVMData{})
}

func GetArguments(pid int) []string {
	if pid == 0 {
		return nil
	}
	filename := "/proc/" + strconv.Itoa(pid) + "/cmdline"
	p := pipe.Line(
		pipe.ReadFile(filename),
		pipe.Exec("strings", "-1"),
	)
	output, err := pipe.CombinedOutput(p)
	if err != nil {
		//fmt.Printf("error:[%v]\n", err)
	}
	return strings.Fields(string(output))
}
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
		log.Fatalf("openToken error: %v\n", err)
	}
	fmt.Println("openToken:", openToken)
	res := make([]interface{}, 0)
	for decoder.More() {
		var m interface{}
		err := decoder.Decode(&m)
		if err != nil {
			log.Fatalf("error decoding: %v\n", err)
		}
		res = append(res, m)
		//fmt.Printf("decoded: [%v]\n", m)
	}
	closeToken, err := decoder.Token()
	if err != nil {
		log.Fatalf("closeToken error: %v\n", err)
	}
	fmt.Println("closeToken:", closeToken)
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
		m := make(map[string]float64)
		for _, e := range j.Res {
			tastd, _ := e.(map[string]interface{})
			procName, _ := tastd["execname"].(string)
			if procName == "qemu-kvm" {
				// fmt.Printf("%v %v\n", i, tastd)
				tpid, _ := tastd["pid"].(float64)
				ipid := int(tpid)
				vmName := GetArguments(ipid)[2]
				ltc, _ := tastd["latency"].(float64)
				m[vmName] += ltc
			}
		}
		fmt.Printf("%v %v\n", len(m), m)
		for i, e := range m {
			if j.LType == LRead {
				vminfo := RawVMData{
					VMName:      i,
					LatencyRead: e / j.ExecPeriod.Seconds(),
				}
				fmt.Printf("%v\n", vminfo)
				DB.Where(RawVMData{VMName: i}).Assign(RawVMData{LatencyRead: e / j.ExecPeriod.Seconds()}).FirstOrCreate(&vminfo)
			} else if j.LType == LWrite {
				vminfo := RawVMData{
					VMName:       i,
					LatencyWrite: e / j.ExecPeriod.Seconds(),
				}
				fmt.Printf("%v\n", vminfo)
				DB.Where(RawVMData{VMName: i}).Assign(RawVMData{LatencyWrite: e / j.ExecPeriod.Seconds()}).FirstOrCreate(&vminfo)

			}
		}
		j.InitCmd()
		j.ExecStreaming()
	}

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
		fmt.Printf("-->%v\n", j.Res)
		for i, e := range j.Res {
			tmp, _ := e.(map[string]interface{})
			s_pid, _ := tmp["pid"].(string)
			var e_pid int
			if len(s_pid) != 0 {
				e_pid, err = strconv.Atoi(s_pid)
				if err != nil {
					log.Fatalf("error convert atoi: %v\n", err)
				}
			}
			argList := GetArguments(e_pid)
			if len(argList) != 0 {
				if argList[0] == "/usr/libexec/qemu-kvm" {
					fmt.Printf("%v %v %v\n", i, tmp, argList[2])
					ioread, _ := tmp["read"].(float64)
					iowrite, _ := tmp["write"].(float64)
					vminfo := RawVMData{
						VMName:    argList[2],
						IOPSRead:  ioread / j.ExecPeriod.Seconds(),
						IOPSWrite: iowrite / j.ExecPeriod.Seconds(),
					}
					fmt.Printf("\n[%v]\n", vminfo)
					//DB.Where(RawVMData{VMName: argList[2]}).Assign(vminfo).FirstOrCreate(&vminfo)
					DB.Where(RawVMData{VMName: argList[2]}).
						Assign(RawVMData{IOPSRead: ioread / j.ExecPeriod.Seconds(), IOPSWrite: iowrite / j.ExecPeriod.Seconds()}).
						FirstOrCreate(&vminfo)
					//DB.Create(&vminfo)
				}
			}
		}
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
