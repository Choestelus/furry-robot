package corgis

import "github.com/jinzhu/gorm"

type RawVMData struct {
	gorm.Model
	VMName       string `gorm:"unique"`
	LatencyRead  float64
	LatencyWrite float64
	IOPSRead     float64
	IOPSWrite    float64
	ISSSD        bool
}

func (RawVMData) TableName() string {
	return "tiramisu_rawdata"
}
