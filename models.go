package corgis

import "github.com/jinzhu/gorm"

type RawVMData struct {
	gorm.Model
	VMName       string  `gorm:"unique"`
	LatencyRead  float64 `gorm:"type:double precision"`
	LatencyWrite float64 `gorm:"type:double precision"`
	IOPSRead     float64 `gorm:"type:double precision"`
	IOPSWrite    float64 `gorm:"type:double precision"`
	ISSSD        bool
}

func (RawVMData) TableName() string {
	return "tiramisu_rawdata"
}
