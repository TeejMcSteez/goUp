package server

import (
	"encoding/json"
	"fmt"
	"goUp/scheduler"
	"goUp/utils"
	"io/fs"
	"net/http"
	"os"
)

var svcData []utils.ServiceData;

var rootFs fs.FS = os.DirFS("server/static")

func Root(w http.ResponseWriter, req *http.Request) {

	path := req.URL.Path

	switch path {
	case "/":
		path = "index.html"
	case "/index.js":
		path = "index.js"
	case "/favicon.ico":
		path = "goUp.png"
	case "/goUp.png":
		path = "goUp.png"
	case "/styles.css":
		path = "styles.css"
	default:
		http.NotFound(w, req)
		return
	}

	http.ServeFileFS(w, req, rootFs, path)
}

func Api(w http.ResponseWriter, req *http.Request) {
    jsonSvcData := make([]utils.ServiceData, 0, len(svcData))

    for i := range svcData {
        jsonSvcData = append(jsonSvcData, utils.ServiceData{
            ServiceName:     svcData[i].ServiceName,
            ServiceHTTPResponse: svcData[i].ServiceHTTPResponse,
			ServiceAPIResponse: svcData[i].ServiceAPIResponse,
        })
    }

    w.Header().Set("Content-Type", "application/json")

    if err := json.NewEncoder(w).Encode(jsonSvcData); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

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
			w.Write([]byte("{ \"updated\": true }"))
		} else {
			w.Header().Add("Content-Type", "application/json")
			http.Error(w, "{ \"updated\": false }", http.StatusBadRequest)
			return
		}
	case "GET":
		if err := json.NewEncoder(w).Encode(ScheduleGetter()); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
		return
    }
	default:
		http.Error(w, "Invalid API Request", http.StatusInternalServerError)
		return
	}
}

func ScheduleUpdater(span int, interval string) bool {
	return scheduler.UpdateParameters(span, interval)
}

func ScheduleGetter() utils.ParamtersData {
	out := scheduler.GetParameters()
	return out
}


func Start(svd []utils.ServiceData) (string, error) {
	svcData = svd
	http.HandleFunc("/", Root)
	http.HandleFunc("/api", Api)
	http.HandleFunc("/api/schedule", ScheduleApi)
	fmt.Println("Starting server at http://localhost:8080/ . . .")
	http.ListenAndServe(":8080", nil)

	return "Started", nil
}