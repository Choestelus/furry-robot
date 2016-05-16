package corgis

import (
	"log"

	"github.com/fatih/color"
)

var updatingFlagSSD bool
var updatingFlagHDD bool

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

func getPostmarkIOPSHDD() (avgIOPS float64) {
	var pVMList []RawVMData
	DB.Model(&RawVMData{}).Where("vm_name = ?", "35monitor_HDD").Find(&pVMList)
	for _, e := range pVMList {
		avgIOPS += e.IOPSRead + e.IOPSWrite
	}
	avgIOPS = (0.5 * avgIOPS) / float64(len(pVMList))
	return
}

func getPostmarkLatencyHDD() (avgLatency float64) {
	var pVMList []RawVMData
	DB.Model(&RawVMData{}).Where("vm_name = ?", "35monitor_HDD").Find(&pVMList)
	for _, e := range pVMList {
		avgLatency += e.LatencyRead + e.LatencyWrite
	}
	avgLatency = (0.5 * avgLatency) / float64(len(pVMList))
	return
}

func getPostmarkIOPSSSD() (avgIOPS float64) {
	var pVMList []RawVMData
	DB.Model(&RawVMData{}).Where("vm_name = ?", "35monitor_SSD").Find(&pVMList)
	for _, e := range pVMList {
		avgIOPS += e.IOPSRead + e.IOPSWrite
	}
	avgIOPS = (0.5 * avgIOPS) / float64(len(pVMList))
	return
}

func getPostmarkLatencySSD() (avgLatency float64) {
	var pVMList []RawVMData
	DB.Model(&RawVMData{}).Where("vm_name = ?", "35monitor_SSD").Find(&pVMList)
	for _, e := range pVMList {
		avgLatency += e.LatencyRead + e.LatencyWrite
	}
	avgLatency = (0.5 * avgLatency) / float64(len(pVMList))
	return
}

func AssignState() {
	//log.Printf("AssignState Invoked\n")
	var HDDVMList []RawVMData
	var SSDVMList []RawVMData
	var AllVMList []RawVMData

	DB.Model(&RawVMData{}).Where("isssd = ?", true).Find(&SSDVMList)
	DB.Model(&RawVMData{}).Where("isssd = ?", false).Find(&HDDVMList)

	var avgHDDLatency float64
	var avgSSDLatency float64
	var avgSSDIOPS float64
	var avgHDDIOPS float64

	//TODO: update IOPS only when postmark is running, update latency when not.
	if updatingFlagHDD == true {
		monitorHDD := &TiramisuState{}
		DB.Model(&TiramisuState{}).Where("vm_name = ?", "35monitor_HDD").First(&monitorHDD)

		avgHDDIOPS = getPostmarkIOPSHDD()
		avgHDDLatency = getPostmarkLatencyHDD()

		log.Printf(color.New(color.FgYellow).Add(color.Bold).SprintfFunc()("realtimeHDD: iops: %v l: %v", avgHDDIOPS, avgHDDLatency))
	} else {
		monitorHDD := &TiramisuState{}
		DB.Model(&TiramisuState{}).Where("vm_name = ?", "35monitor_HDD").First(&monitorHDD)

		avgHDDIOPS = monitorHDD.IOPS
		avgHDDLatency = getPostmarkLatencyHDD()

		log.Printf(color.New(color.FgYellow).Add(color.Bold).SprintfFunc()("passiveHDD: iops: %v l: %v", avgHDDIOPS, avgHDDLatency))
	}

	if updatingFlagSSD == true {
		monitorSSD := &TiramisuState{}
		DB.Model(&TiramisuState{}).Where("vm_name = ?", "35monitor_SSD").First(&monitorSSD)

		avgSSDIOPS = getPostmarkIOPSSSD()
		avgSSDLatency = getPostmarkLatencySSD()

		log.Printf(color.New(color.FgYellow).Add(color.Bold).SprintfFunc()("realtimeSSD: iops: %v l: %v", avgSSDIOPS, avgSSDLatency))
	} else {
		monitorSSD := &TiramisuState{}
		DB.Model(&TiramisuState{}).Where("vm_name = ?", "35monitor_SSD").First(&monitorSSD)

		avgSSDIOPS = monitorSSD.IOPS
		avgSSDLatency = getPostmarkLatencySSD()

		log.Printf(color.New(color.FgYellow).Add(color.Bold).SprintfFunc()("passiveSSD: iops: %v l: %v", avgSSDIOPS, avgSSDLatency))
	}

	// avgHDDIOPS = calcAvgIOPS(false, &HDDVMList)
	// avgHDDLatency = calcAvgLatency(false, &HDDVMList)
	// avgSSDIOPS = calcAvgIOPS(true, &SSDVMList)
	// avgSSDLatency = calcAvgLatency(true, &SSDVMList)

	AllVMList = append(AllVMList, HDDVMList...)
	AllVMList = append(AllVMList, SSDVMList...)
	for _, e := range AllVMList {
		tiramisu_state := TiramisuState{
			IOPS:        e.IOPSRead + e.IOPSWrite,
			Latency:     (e.LatencyRead + e.LatencyWrite) / 2,
			IOPS_HDD:    avgHDDIOPS,
			Latency_HDD: avgHDDLatency,
			IOPS_SSD:    avgSSDIOPS,
			Latency_SSD: avgSSDLatency,
			Name:        e.VMName,
		}
		greenf := color.New(color.FgGreen).Add(color.Bold).SprintfFunc()
		log.Printf(greenf("Name: %v IS: %v LS: %v IH: %v LH: %v", tiramisu_state.Name,
			// 	// tiramisu_state.IOPS, tiramisu_state.Latency,
			tiramisu_state.IOPS_SSD, tiramisu_state.Latency_SSD,
			tiramisu_state.IOPS_HDD, tiramisu_state.Latency_HDD,
		))
		DB.Model(&TiramisuState{}).Where("vm_name = ?", e.VMName).Assign(tiramisu_state).FirstOrCreate(&tiramisu_state)
	}

	// for _, e := range HDDVMList {
	// 	tiramisu_state := TiramisuState{
	// 		IOPS:        e.IOPSRead + e.IOPSWrite,
	// 		Latency:     (e.LatencyRead + e.LatencyWrite) / 2,
	// 		IOPS_HDD:    e.IOPSRead + e.IOPSWrite,
	// 		Latency_HDD: (e.LatencyRead + e.LatencyWrite) / 2,
	// 		IOPS_SSD:    avgSSDIOPS,
	// 		Latency_SSD: avgSSDLatency,
	// 		Name:        e.VMName,
	// 	}
	// 	DB.Model(&TiramisuState{}).Where("vm_name = ?", e.VMName).Assign(tiramisu_state).FirstOrCreate(&tiramisu_state)
	// }
	// for _, e := range SSDVMList {
	// 	tiramisu_state := TiramisuState{
	// 		IOPS:        (e.IOPSRead + e.IOPSWrite),
	// 		Latency:     (e.LatencyRead + e.LatencyWrite) / 2,
	// 		IOPS_HDD:    avgHDDIOPS,
	// 		Latency_HDD: avgHDDLatency,
	// 		IOPS_SSD:    (e.IOPSRead + e.IOPSWrite),
	// 		Latency_SSD: (e.LatencyRead + e.LatencyWrite) / 2,
	// 		Name:        e.VMName,
	// 	}
	// 	DB.Model(&TiramisuState{}).Where("vm_name = ?", e.VMName).Assign(tiramisu_state).FirstOrCreate(&tiramisu_state)
	// }
}
