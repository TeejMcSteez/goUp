package utils

import (
	"goUp/utils"
	"testing"
)

// Testing getting recent service data in different threads to make sure their is no SQL thread error
func TestRecentDataSQLFail(t *testing.T) {
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