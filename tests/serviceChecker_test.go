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