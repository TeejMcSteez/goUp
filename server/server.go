package server

import (
	"encoding/json"
	"fmt"
	"goUp/scheduler"
	"goUp/utils"
	"io/fs"
	"net/http"
	"embed"
)

var svcData *[]utils.ServiceData
var scd *scheduler.Scheduler

//go:embed static/index.html static/index.js static/styles.css static/goUp.png
var content embed.FS
var rootFs fs.FS = content

// Client HTML server
func Root(w http.ResponseWriter, req *http.Request) {

	path := req.URL.Path

	switch path {
	case "/":
		path = "static/index.html"
	case "/index.js":
		path = "static/index.js"
	case "/favicon.ico":
		path = "static/goUp.png"
	case "/goUp.png":
		path = "static/goUp.png"
	case "/styles.css":
		path = "static/styles.css"
	default:
		http.NotFound(w, req)
		return
	}

	http.ServeFileFS(w, req, rootFs, path)
}

// Basic API server, returns current service data
func Api(w http.ResponseWriter, req *http.Request) {
	data := *svcData
	jsonSvcData := make([]utils.ServiceData, 0, len(data))
	for i := range data {
		jsonSvcData = append(jsonSvcData, utils.ServiceData{
			ServiceName:         data[i].ServiceName,
			ServiceHTTPResponse: data[i].ServiceHTTPResponse,
			ServiceAPIResponse:  data[i].ServiceAPIResponse,
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
		var state scheduler.ScheduleState = scd.Get()
		if _, err := w.Write([]byte("{ \"timespan\":" + fmt.Sprint(state.Span) +", \"timeInterval\": \"" + state.Interval + "\" }")); err != nil {
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

// Starts server with all handler functions
func Start(svd *[]utils.ServiceData, sch *scheduler.Scheduler) (string, error) {
	svcData = svd
	scd = sch
	http.HandleFunc("/", Root)
	http.HandleFunc("/api", Api)
	http.HandleFunc("/api/schedule", ScheduleApi)
	http.HandleFunc("/api/status", StatusApi)
	fmt.Println("Starting server at http://localhost:8101/ . . .")
	if err := http.ListenAndServe(":8101", nil); err != nil {
		panic(err)
	}

	return "Started", nil
}
