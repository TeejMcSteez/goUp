package server

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"goUp/scheduler"
	"goUp/utils"
	"io/fs"
	"net/http"
)

var db *sql.DB
var scd *scheduler.Scheduler

//go:embed all:static
var content embed.FS

// Basic API server, returns current service data
func Api(w http.ResponseWriter, req *http.Request) {
	data := utils.GetRecentData(db)
	jsonSvcData := make([]utils.ServiceData, 0, len(data))
	for i := range data {
		jsonSvcData = append(jsonSvcData, utils.ServiceData{
			ServiceName:         data[i].ServiceName,
			ServiceHTTPResponse: data[i].ServiceHTTPResponse,
			ServiceAPIResponse:  data[i].ServiceAPIResponse,
			ServiceResponseTime: data[i].ServiceResponseTime,
		})
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(jsonSvcData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Schedule API server, returns current schedule parameters
func ScheduleApi(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		dec := json.NewDecoder(req.Body)
		var jsonData utils.ParamtersData

		if err := dec.Decode(&jsonData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		updated := ScheduleUpdater(jsonData.Span, jsonData.Interval)

		if updated {
			w.Header().Add("Content-Type", "application/json")
			if _, err := w.Write([]byte("{ \"updated\": true }")); err != nil {
				panic(err)
			}
		} else {
			w.Header().Add("Content-Type", "application/json")
			http.Error(w, "{ \"updated\": false }", http.StatusBadRequest)
			return
		}
	case "GET":
		// Add Get case in scheduler
		w.Header().Add("Content-Type", "application/json")
		state := scd.Get()
		if _, err := w.Write([]byte("{ \"timespan\":" + fmt.Sprint(state.Span) + ", \"timeInterval\": \"" + state.Interval + "\" }")); err != nil {
			panic(err)
		}
	default:
		http.Error(w, "Invalid API Request", http.StatusBadRequest)
		return
	}
}

// Updates schedule parameters
func ScheduleUpdater(span int, interval string) bool {
	ok := scd.Update(span, interval)

	return ok
}

// Gets current status in JSON format for automated fetching
func StatusApi(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		http.Error(w, "Invalid method", http.StatusBadRequest)
	case "GET":
		w.Header().Add("Content-Type", "application/json")
		apiData := utils.GetServiceData()
		if err := json.NewEncoder(w).Encode(apiData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Bad request", http.StatusInternalServerError)
		return
	}
}

func UptimeAPI(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		http.Error(w, "Invalid method", http.StatusBadRequest)
	case "GET":
		w.Header().Add("Content-Type", "application/json")
		endpoints := utils.GetServiceEndpoints()
		var avgData []utils.AverageData
		for idx := range endpoints {
			endpointName := endpoints[idx].URL
			avgData = append(avgData, utils.AverageData{Name: endpointName, Average: utils.GetUptimeAverages(db, endpointName)})
		}
		if err := json.NewEncoder(w).Encode(avgData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Bad request", http.StatusInternalServerError)
		return
	}
}

// Starts server with all handler functions
func Start(database *sql.DB, sch *scheduler.Scheduler) error {
	db = database
	scd = sch

	staticFS, err := fs.Sub(content, "static")
	if err != nil {
		return err
	}

	http.Handle("/", http.FileServer(http.FS(staticFS)))
	http.HandleFunc("/api", Api)
	http.HandleFunc("/api/schedule", ScheduleApi)
	http.HandleFunc("/api/status", StatusApi)
	http.HandleFunc("/api/uptime", UptimeAPI)
	fmt.Println("Starting server at http://localhost:8101/ . . .")
	if err := http.ListenAndServe(":8101", nil); err != nil {
		return err
	}

	return nil
}
