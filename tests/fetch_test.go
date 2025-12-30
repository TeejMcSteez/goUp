package utils_test

import (
	"goUp/utils"
	"testing"
	"log"
)

func FetchTest(t *testing.T) {
	_, err := utils.GetServiceData()
	if err != nil {
		log.Printf("error getting service data in fetch test: %v", err)
		t.Fail()
	}
}