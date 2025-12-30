package utils_test

import (
	"database/sql"
	"goUp/utils"
	"log"
	"sync"
	"testing"
)

func testWorker(d *sql.DB, t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	for range 30 {
		_, err := utils.GetRecentData(d)
		if err != nil {
			t.Log("Failed getting recent data from the database: ", err)
			t.Fail()
		}
	}
}

// Testing getting recent service data in different threads to make sure their is no SQL thread error
func TestRecentDataSQLFail(t *testing.T) {
	var wg sync.WaitGroup
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

	numWorkers := 4
	wg.Add(numWorkers)

	for range numWorkers {
		go testWorker(d, t, &wg)
	}

	wg.Wait()
	
	if err := d.Close(); err != nil {
		log.Printf("Error closing database testing concurrency: %v", err)
		t.Fail()
	}

}