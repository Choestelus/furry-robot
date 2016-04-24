package corgis

import "log"

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
