package main

import (
	"fmt"
	"goUp/scheduler"
	"goUp/server"
	"goUp/utils"
	"slices"
)

var serviceEndpoints []string = []string{}



func main() {
	fmt.Print("Loading Config . . .\n\n")
	cfg, err := utils.LoadConfig("services.yml")

	if err != nil {
		panic(err)
	}
	for name, svc := range cfg.Services {
		if !slices.Contains(serviceEndpoints, svc.URL) {
			fmt.Println("Adding ", name, "to service endpoints")
			serviceEndpoints = append(serviceEndpoints, svc.URL)
		}
	}
	// Get current service data before full launch
	var svcData = utils.GetServiceData(serviceEndpoints)
	// Create blank data store
	var dataStore utils.SharedData = *utils.NewStore()
	// Set current data to most recent fetch
	dataStore.Set(svcData)
	// Start's scheduler
	scheduler.StartScheduler(serviceEndpoints, &dataStore)
	// Starts http server
	ret, err := server.Start(svcData)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(ret)
	}

}