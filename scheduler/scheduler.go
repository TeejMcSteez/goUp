package scheduler

import (
	"goUp/utils"
	"time"
	"fmt"
)
/*
 Start's a goroutine with a ticker which every 30 seconds gets and updates service data in main
*/
func StartScheduler(serviceEndpoints []utils.Service, currData *utils.SharedData) {
    fmt.Println("Starting service data ticker")
    ticker := time.NewTicker(30 * time.Second)

    go func() {
        for range ticker.C {
            data := utils.GetServiceData(serviceEndpoints)
            currData.Set(data)
            fmt.Println("Fetched data")
        }
    }()
}
