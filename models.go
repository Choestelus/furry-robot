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

type TiramisuStorage struct {
	VMName          string `gorm:"primary_key;type:varchar(100)"`
	CurrentPool     string `gorm:"type:varchar(10)"`
	AppropriatePool string `gorm:"type:varchar(10)";column:appropiate_pool`
	Notice          int    `gorm:"type:integer"`
}

func (TiramisuStorage) TableName() string {
	return "tiramisu_storage"
}

type TiramisuState struct {
	Name        string  `gorm:"column:vm_name;primary_key"`
	Latency     float64 `gorm:"column:latency;type:double precision"`
	IOPS        float64 `gorm:"column:iops;type:double precision"`
	Latency_HDD float64 `gorm:"column:latency_hdd;type:double precision"`
	IOPS_HDD    float64 `gorm:"column:iops_hdd;type:double precision"`
	Latency_SSD float64 `gorm:"column:latency_ssd;type:double precision"`
	IOPS_SSD    float64 `gorm:"column:iops_ssd;type:double precision"`
}

func (TiramisuState) TableName() string {
	return "tiramisu_state"
}
