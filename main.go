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
	db := utils.InitDB()
	// Adds recently fetched data to the database
	for data := range svcData {
		utils.InsertData(db, svcData[data])
	}
	// For now keep in memory store but moving toward db only store
	// Keeping it because I will also need to re-write logic of scheduler to insert data . . .
	// instead of using shared in-memory data
	dataStore := utils.NewStore()
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
