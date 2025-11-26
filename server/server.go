package server

import (
	"fmt"
	"goUp/utils"
	"io/fs"
	"net/http"
	"os"
	"encoding/json"
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


func Start(svd []utils.ServiceData) (string, error) {
	svcData = svd
	http.HandleFunc("/", Root)
	http.HandleFunc("/api", Api)
	fmt.Println("Starting server at http://localhost:8080/ . . .")
	http.ListenAndServe(":8080", nil)

	return "Started", nil
}