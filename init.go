package corgis

import (
	"log"

	"github.com/jinzhu/gorm"
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
	DB.LogMode(false)
	DB.AutoMigrate(&RawVMData{})
}
