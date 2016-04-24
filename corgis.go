package corgis

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gopkg.in/pipe.v2"
)

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
	//log.Printf("timedSIGTERM called\n")
	var _ = <-time.After(d)
	err := p.Signal(syscall.SIGTERM)
	//log.Printf("SIGTERM sent\n")
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
	//fmt.Println("openToken:", openToken)
	var _ = openToken
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
	//fmt.Println("closeToken:", closeToken)
	var _ = closeToken
	return res
}

func (j *JobScheduler) InitCmd() {
	j.Cmd = exec.Command(j.Cmd.Path, j.Cmd.Args[1])
	j.OutBuf.Reset()
	j.ErrBuf.Reset()
	j.Cmd.Stdout = &j.OutBuf
	j.Cmd.Stderr = &j.ErrBuf
	//log.Printf("Initialized job\n")
}

// Curently uses for read&write latencies
func (j *JobScheduler) ExecStreaming() {
	err := j.Cmd.Start()
	if err != nil {
		log.Panicf("cmd start error: %v\n", err)
	}
	//log.Printf("job started\n")
	go timedSIGTERM(j.Cmd.Process, j.ExecPeriod)
	//log.Printf("waiting for job terminate\n")
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
		//fmt.Printf("%v %v\n", len(m), m)
		for i, e := range m {
			if j.LType == LRead {

				var storageType TiramisuStorage
				DB.First(&storageType, "vm_name = ?", i)

				vminfo := RawVMData{
					VMName:      i,
					LatencyRead: e / j.ExecPeriod.Seconds(),
					ISSSD:       storageType.CurrentPool == "SSD",
				}
				//fmt.Printf("%v\n", vminfo)
				DB.Where(RawVMData{VMName: i}).Assign(vminfo).FirstOrCreate(&vminfo)
			} else if j.LType == LWrite {

				var storageType TiramisuStorage
				DB.First(&storageType, "vm_name = ?", i)

				vminfo := RawVMData{
					VMName:       i,
					LatencyWrite: e / j.ExecPeriod.Seconds(),
					ISSSD:        storageType.CurrentPool == "SSD",
				}
				//fmt.Printf("%v\n", vminfo)
				DB.Where(RawVMData{VMName: i}).Assign(vminfo).FirstOrCreate(&vminfo)

			}
		}
		j.InitCmd()
		j.ExecStreaming()
	}

}

// Currently uses for IOPS
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
		//fmt.Printf("-->%v\n", j.Res)
		for _, e := range j.Res {
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
					//fmt.Printf("%v %v %v\n", i, tmp, argList[2])
					ioread, _ := tmp["read"].(float64)
					iowrite, _ := tmp["write"].(float64)

					var storageType TiramisuStorage
					DB.First(&storageType, "vm_name = ?", argList[2])
					// fmt.Printf("\n\n---->[%v]\n\n", storageType)

					vminfo := RawVMData{
						VMName:    argList[2],
						IOPSRead:  ioread / j.ExecPeriod.Seconds(),
						IOPSWrite: iowrite / j.ExecPeriod.Seconds(),
						ISSSD:     storageType.CurrentPool == "SSD",
					}
					//fmt.Printf("\n\n---->[%v]\n\n", vminfo)
					//fmt.Printf("\n[%v]\n", vminfo)
					DB.Where(RawVMData{VMName: argList[2]}).Assign(vminfo).FirstOrCreate(&vminfo)
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

// TODO: Refactor avg parameters as module.

func calcAvgIOPS(isssd bool, pVMList *[]RawVMData) (avgIOPS float64) {
	DB.Model(&RawVMData{}).Where("isssd = ?", isssd).Find(pVMList)
	for _, e := range *pVMList {
		avgIOPS += e.IOPSRead + e.IOPSWrite
	}
	avgIOPS = (0.5 * avgIOPS) / float64(len(*pVMList))

	return
}

func calcAvgLatency(isssd bool, pVMList *[]RawVMData) (avgLatency float64) {
	DB.Model(&RawVMData{}).Where("isssd = ?", isssd).Find(pVMList)
	for _, e := range *pVMList {
		avgLatency += e.LatencyRead + e.LatencyWrite
	}
	avgLatency = (0.5 * avgLatency) / float64(len(*pVMList))
	return
}

func AssignAverage() {
	log.Printf("AssignAverage Invoked\n")
	var HDDVMList []RawVMData
	var SSDVMList []RawVMData
	DB.Model(&RawVMData{}).Where("isssd = ?", true).Find(&SSDVMList)
	DB.Model(&RawVMData{}).Where("isssd = ?", false).Find(&HDDVMList)

	var avgHDDLatency float64
	var avgSSDLatency float64
	var avgSSDIOPS float64
	var avgHDDIOPS float64

	for _, e := range HDDVMList {
		avgHDDIOPS += e.IOPSRead + e.IOPSWrite
		avgHDDLatency += e.LatencyRead + e.LatencyWrite
	}
	for _, e := range SSDVMList {
		avgSSDIOPS += e.IOPSRead + e.IOPSWrite
		avgSSDLatency += e.LatencyRead + e.LatencyWrite
	}
	avgHDDIOPS = calcAvgIOPS(false, &HDDVMList)
	avgHDDLatency = calcAvgLatency(false, &HDDVMList)
	avgSSDIOPS = calcAvgIOPS(true, &SSDVMList)
	avgSSDLatency = calcAvgLatency(true, &SSDVMList)

	for _, e := range HDDVMList {
		tiramisu_state := TiramisuState{
			IOPS:        (e.IOPSRead + e.IOPSWrite) / 2,
			Latency:     (e.LatencyRead + e.LatencyWrite) / 2,
			IOPS_HDD:    (e.IOPSRead + e.IOPSWrite) / 2,
			Latency_HDD: (e.LatencyRead + e.LatencyWrite) / 2,
			IOPS_SSD:    avgSSDIOPS,
			Latency_SSD: avgSSDLatency,
			Name:        e.VMName,
		}
		DB.Model(&TiramisuState{}).Where("vm_name = ?", e.VMName).Assign(tiramisu_state).FirstOrCreate(&tiramisu_state)
	}
	for _, e := range SSDVMList {
		tiramisu_state := TiramisuState{
			IOPS:        (e.IOPSRead + e.IOPSWrite) / 2,
			Latency:     (e.LatencyRead + e.LatencyWrite) / 2,
			IOPS_HDD:    avgHDDIOPS,
			Latency_HDD: avgHDDLatency,
			IOPS_SSD:    (e.IOPSRead + e.IOPSWrite) / 2,
			Latency_SSD: (e.LatencyRead + e.LatencyWrite) / 2,
			Name:        e.VMName,
		}
		DB.Model(&TiramisuState{}).Where("vm_name = ?", e.VMName).Assign(tiramisu_state).FirstOrCreate(&tiramisu_state)
	}
}
