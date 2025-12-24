package utils_test

import (
	"goUp/utils"
	"testing"
)

func TestServiceCheck(t *testing.T) {
	sd := []utils.ServiceData{
		{ServiceName: "test1", ServiceHTTPResponse: "200", ServiceAPIResponse: "", ServiceResponseTime: ""},
		{ServiceName: "test2", ServiceHTTPResponse: "303", ServiceAPIResponse: "", ServiceResponseTime: ""},
	}

	res, err := utils.Check(sd)

	if err != nil {
		t.Error("Error checking data")
	}

	if len(res) > 1 {
		t.Error("Should only be one error")
	}
}