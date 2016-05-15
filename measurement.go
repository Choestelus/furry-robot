package corgis

import "log"

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

func getPostmarkIOPSHDD(pVMList *[]RawVMData) (avgIOPS float64) {
	DB.Model(&RawVMData{}).Where("vm_name = ?", "35monitor_HDD").Find(pVMList)
	for _, e := range *pVMList {
		avgIOPS += e.IOPSRead + e.IOPSWrite
	}
	avgIOPS = (0.5 * avgIOPS) / float64(len(*pVMList))
	return
}

func getPostmarkLatencyHDD(pVMList *[]RawVMData) (avgLatency float64) {
	DB.Model(&RawVMData{}).Where("vm_name = ?", "35monitor_HDD").Find(pVMList)
	for _, e := range *pVMList {
		avgLatency += e.LatencyRead + e.LatencyWrite
	}
	avgLatency = (0.5 * avgLatency) / float64(len(*pVMList))
	return
}

func getPostmarkIOPSSSD(pVMList *[]RawVMData) (avgIOPS float64) {
	DB.Model(&RawVMData{}).Where("vm_name = ?", "35monitor_SSD").Find(pVMList)
	for _, e := range *pVMList {
		avgIOPS += e.IOPSRead + e.IOPSWrite
	}
	avgIOPS = (0.5 * avgIOPS) / float64(len(*pVMList))
	return
}

func getPostmarkLatencySSD(pVMList *[]RawVMData) (avgLatency float64) {
	DB.Model(&RawVMData{}).Where("vm_name = ?", "35monitor_SSD").Find(pVMList)
	for _, e := range *pVMList {
		avgLatency += e.LatencyRead + e.LatencyWrite
	}
	avgLatency = (0.5 * avgLatency) / float64(len(*pVMList))
	return
}

func AssignState() {
	log.Printf("AssignState Invoked\n")
	var HDDVMList []RawVMData
	var SSDVMList []RawVMData
	DB.Model(&RawVMData{}).Where("isssd = ?", true).Find(&SSDVMList)
	DB.Model(&RawVMData{}).Where("isssd = ?", false).Find(&HDDVMList)

	var avgHDDLatency float64
	var avgSSDLatency float64
	var avgSSDIOPS float64
	var avgHDDIOPS float64

	//TODO: update only when postmark is running
	if updatingFlagHDD == true {
		avgHDDIOPS = getPostmarkIOPSHDD(&HDDVMList)
		avgHDDLatency = getPostmarkLatencyHDD(&HDDVMList)
	} else {
		monitorHDD := &RawVMData{}
		DB.Model(&RawVMData{}).Where("vm_name = ?", "35monitor_HDD").First(&monitorHDD)

		avgHDDIOPS = monitorHDD.IOPSRead + monitorHDD.IOPSWrite
		avgHDDLatency = (monitorHDD.LatencyRead + monitorHDD.LatencyWrite) * 0.5
	}

	if updatingFlagSSD == true {
		avgSSDIOPS = getPostmarkIOPSSSD(&SSDVMList)
		avgSSDLatency = getPostmarkLatencySSD(&SSDVMList)
	} else {
		monitorSSD := &RawVMData{}
		DB.Model(&RawVMData{}).Where("vm_name = ?", "35monitor_SSD").First(&monitorSSD)

		avgSSDIOPS = monitorSSD.IOPSRead + monitorSSD.IOPSWrite
		avgSSDIOPS = (monitorSSD.LatencyRead + monitorSSD.LatencyWrite) * 0.5
	}
	// avgHDDIOPS = calcAvgIOPS(false, &HDDVMList)
	// avgHDDLatency = calcAvgLatency(false, &HDDVMList)
	// avgSSDIOPS = calcAvgIOPS(true, &SSDVMList)
	// avgSSDLatency = calcAvgLatency(true, &SSDVMList)

	for _, e := range HDDVMList {
		tiramisu_state := TiramisuState{
			IOPS:        e.IOPSRead + e.IOPSWrite,
			Latency:     (e.LatencyRead + e.LatencyWrite) / 2,
			IOPS_HDD:    e.IOPSRead + e.IOPSWrite,
			Latency_HDD: (e.LatencyRead + e.LatencyWrite) / 2,
			IOPS_SSD:    avgSSDIOPS,
			Latency_SSD: avgSSDLatency,
			Name:        e.VMName,
		}
		DB.Model(&TiramisuState{}).Where("vm_name = ?", e.VMName).Assign(tiramisu_state).FirstOrCreate(&tiramisu_state)
	}
	for _, e := range SSDVMList {
		tiramisu_state := TiramisuState{
			IOPS:        (e.IOPSRead + e.IOPSWrite),
			Latency:     (e.LatencyRead + e.LatencyWrite) / 2,
			IOPS_HDD:    avgHDDIOPS,
			Latency_HDD: avgHDDLatency,
			IOPS_SSD:    (e.IOPSRead + e.IOPSWrite),
			Latency_SSD: (e.LatencyRead + e.LatencyWrite) / 2,
			Name:        e.VMName,
		}
		DB.Model(&TiramisuState{}).Where("vm_name = ?", e.VMName).Assign(tiramisu_state).FirstOrCreate(&tiramisu_state)
	}
}
