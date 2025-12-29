package utils

import (
	"goUp/utils"
	"log"
	"testing"
)

// Testing getting recent service data in different threads to make sure their is no SQL thread error
func TestRecentDataSQLFail(t *testing.T) {
	cfg, err := utils.LoadConfig("../services.yml")
	if err != nil {
		log.Printf("Error occured testing SQL data failure")
	}
	utils.Current_Config = cfg
	d, err := utils.InitDB()
	if err != nil {
		t.Log("Error initializing the database: ", err)
		t.Fail()
	}
	go func(tes *testing.T) {
		for range 30 {
			_, err := utils.GetRecentData(d)
			if err != nil {
				t.Log("Failed getting recent data from the database: ", err)
				t.Fail()
			}
		}
	}(t)
	go func(tes *testing.T) {
		for range 30 {
			_, err := utils.GetRecentData(d)
			if err != nil {
				t.Log("Failed getting recent data from the database: ", err)
				t.Fail()
			}
		}
	}(t)

}