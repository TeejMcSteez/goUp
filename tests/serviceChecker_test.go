package utils_test

import (
	"goUp/utils"
	"testing"
	"log"
)

func TestServiceCheck(t *testing.T) {
	cfg, err := utils.LoadConfig("../services.yml")
	if err != nil {
		log.Printf("Error occured loading config: %v", err)
		t.Fail()
	}
	sd := []utils.ServiceData{
		{ServiceName: "test1", ServiceHTTPResponse: "200", ServiceAPIResponse: "", ServiceResponseTime: ""},
		{ServiceName: "test2", ServiceHTTPResponse: "303", ServiceAPIResponse: "", ServiceResponseTime: ""},
	}
	utils.Current_Config = cfg
	res, err := utils.Check(sd)

	if err != nil {
		log.Printf("%v", err)
		t.Error("Error checking data")
	}

	if len(res) > 1 {
		t.Error("Should only be one error")
	}
}

func TestUptimeCalculation(t *testing.T) {
	cfg, err := utils.LoadConfig("../services.yml")
	if err != nil {
		log.Printf("Failed to load configuration while calculating load time averages")
		t.Fail()
	}
	if cfg != nil {
		utils.Current_Config = cfg
	} else {
		log.Print("Current configuration is currently null check path!")
	}
	db, err := utils.InitDB()
	if err != nil {
		log.Printf("Failed with error: %v", err)
		t.Fail()
	}
	avgs, err := utils.GetUptimeAverage(db, "n8n")
	if err != nil {
		log.Printf("Error occured: %v", err)
		t.Fail()
	} else {
		log.Printf("Got average from serviceChecker: %v", avgs)
	}
}