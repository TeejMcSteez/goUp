package scheduler

import (
	"fmt"
	"goUp/utils"
	"strings"
	"sync"
	"time"
)

// Default Timespan:
var s int = 30
var i string = "Seconds"
var mux sync.RWMutex
var schedule utils.ScheduleParameters = utils.ScheduleParameters{Span: &s, Interval: &i, Mux: &mux}

var ticker time.Ticker

/*
 Start's a goroutine with a ticker which every 30 seconds gets and updates service data in main
*/
func StartScheduler(serviceEndpoints []utils.Service, currData *utils.SharedData) {
    fmt.Println("Starting service data ticker")
    span := time.Duration(*schedule.Span)
    var dur time.Duration
    switch strings.ToLower(i)[0] {
    case 's':
        dur = span * time.Second
    case 'h':
        dur = span * time.Hour
    case 'm':
        dur = span * time.Minute
    default:
        panic("Invalid time input")
    }

    ticker = *time.NewTicker(dur)
    go func() {
        for range ticker.C {
            data := utils.GetServiceData(serviceEndpoints)
            currData.Set(data)
            fmt.Println("Fetched data")
        }
    }()
}

func UpdateParamters(span int, interval string) {
    fmt.Println("Stopping ticker")
    ticker.Stop()
    fmt.Println("Updating ticker parameters")
    schedule.Mux.Lock()

    defer schedule.Mux.Unlock()

    *schedule.Span = span
    *schedule.Interval =interval

    fmt.Println("Schedule update completed")
}

func GetParameters() utils.ParamtersData {
    schedule.Mux.Lock()

    defer schedule.Mux.Unlock()

    out := utils.ParamtersData{ Span: *schedule.Span, Interval: *schedule.Interval }

    return out
}