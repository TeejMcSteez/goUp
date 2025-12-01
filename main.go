package main

import (
	"fmt"
	"goUp/scheduler"
	"goUp/server"
	"goUp/utils"
)

func main() {
	// Get current service data before full launch
	svcData := utils.GetServiceData()
	// Create blank data store
	dataStore := utils.NewStore()
	// Set current data to most recent fetch
	dataStore.Set(svcData)
	// Starts scheduler
	sch := scheduler.NewScheduler(dataStore, 30, "seconds")
	defer sch.Stop()

	// Starts http server
	ret, err := server.Start(&svcData, sch)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(ret)
	}

}
