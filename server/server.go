package server

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"goUp/utils"
	scheduler "goUp/workers"
	"io/fs"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	db  *sql.DB
	scd *scheduler.Scheduler
}

//go:embed all:static
var content embed.FS

// API handler, returns current service data from the database
func (s *Server) Api(w http.ResponseWriter, req *http.Request) {
	data, err := utils.GetRecentData(s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Schedule API handler
//
// POST: Will try to parse request body as JSON and update the current schedule, will respond with a boolean.
//
// GET: Gets current schedule information and returns as JSON
//
// Any other method gets a bad request response
func (s *Server) ScheduleApi(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		dec := json.NewDecoder(req.Body)
		var jsonData scheduler.ScheduleState

		if err := dec.Decode(&jsonData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		updated := s.scheduleUpdater(jsonData)

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
		state := s.scd.Get()
		if err := json.NewEncoder(w).Encode(state); err != nil {
			panic(err)
		}
	default:
		http.Error(w, "Invalid API Request", http.StatusBadRequest)
		return
	}
}

// Updates schedule parameters from schedule API
func (s *Server) scheduleUpdater(state scheduler.ScheduleState) bool {
	ok := s.scd.Update(state)

	return ok
}

// Gets current status in JSON format for automated fetching
//
// Only accepts GET requests which will return currently down services
func (s *Server) StatusApi(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		http.Error(w, "Invalid method", http.StatusBadRequest)
	case "GET":
		w.Header().Add("Content-Type", "application/json")

		recData, err := utils.GetRecentData(s.db)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		apiData, err := utils.Check(recData)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if err := json.NewEncoder(w).Encode(apiData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Bad request", http.StatusInternalServerError)
		return
	}
}

// Gets current uptime averages from the backend
func (s *Server) UptimeAPI(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		http.Error(w, "Invalid method", http.StatusBadRequest)
	case "GET":
		w.Header().Add("Content-Type", "application/json")
		endpoints := utils.GetServiceEndpoints()
		var avgData []utils.AverageData
		for idx := range endpoints {
			endpointName := endpoints[idx].Name
			upAvg, err := utils.GetUptimeAverage(s.db, endpointName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			avgData = append(avgData, utils.AverageData{Name: endpointName, Average: upAvg})
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

// Returns size of database in bytes
func (s *Server) GetDatabaseSize(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	size, err := utils.GetDatabaseSize()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err = json.NewEncoder(w).Encode(&utils.DatabaseSizePayload{Size: size}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Clears database
func (s *Server) ClearDatabase(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if err := utils.ClearDatabase(s.db); err != nil {
		json.NewEncoder(w).Encode([]byte(err.Error()))
	}
	fmt.Fprint(w, `{ "ok": true }`)
}

// Returns a new server instance
func NewServer(db *sql.DB, scd *scheduler.Scheduler) *Server {
	return &Server{db: db, scd: scd}
}

// Gets all service data which has errored
func (s *Server) GetErrorData(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	param := req.URL.Query().Get("limit")
	limit, err := strconv.Atoi(param)
	if err != nil {
		log.Printf("Error occured, invalid limit: %v\nError: %v", limit, err)
		limit = 0
	}
	sortOrder := req.URL.Query().Get("sort")
	data, err := utils.GetErrorData(s.db, limit, sortOrder)
	if err != nil {
		log.Printf("Error occured getting error data from database: %v", err)
		json.NewEncoder(w).Encode([]byte(err.Error()))
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding error data to json: %v", err)
		json.NewEncoder(w).Encode([]byte(err.Error()))
		return
	}
}

func (s *Server) ConfigServiceApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		var service utils.Service
		if err := json.NewDecoder(req.Body).Decode(&service); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := utils.AddConfigService(utils.Current_Config, service); err != nil {
			log.Printf("Error adding config service: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, `{ "ok": true }`)
	case "DELETE":
		var service utils.Service
		if err := json.NewDecoder(req.Body).Decode(&service); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := utils.DeleteConfigService(utils.Current_Config, service); err != nil {
			log.Printf("Error deleting config service: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, `{ "ok": true }`)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

func (s *Server) ConfigMQTTApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		var mqtt utils.MQTTTrigger
		if err := json.NewDecoder(req.Body).Decode(&mqtt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := utils.AddConfigMQTTTrigger(utils.Current_Config, mqtt); err != nil {
			log.Printf("Error adding MQTT config: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, `{ "ok": true }`)
	case "DELETE":
		if err := utils.DeleteConfigMQTT(utils.Current_Config); err != nil {
			log.Printf("Error deleting MQTT config: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, `{ "ok": true }`)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

func (s *Server) ConfigWebhookApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		var webhook utils.WebhookTrigger
		if err := json.NewDecoder(req.Body).Decode(&webhook); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := utils.AddConfigWebhookTrigger(utils.Current_Config, webhook); err != nil {
			log.Printf("Error adding webhook config: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, `{ "ok": true }`)
	case "DELETE":
		if err := utils.DeleteConfigTrigger(utils.Current_Config); err != nil {
			log.Printf("Error deleting webhook config: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, `{ "ok": true }`)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

func (s *Server) ReadConfigData(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sData := utils.ReadConfigServices(utils.Current_Config)
	mData := utils.ReadConfigMQTT(utils.Current_Config)
	wData := utils.ReadConfigWebhook(utils.Current_Config)

	cData := utils.ConfigData{Services: sData, MQTT: mData, Webhook: wData}

	data, err := json.Marshal(cData)
	if err != nil {
		log.Printf("Error occured marshalling JSON data when reading configuration information: %v", err)
		http.Error(w, "Server Error: Failed to parse config data", 500)
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding configuration data to json: %v", err)
		http.Error(w, "Server Error: Failed to parse config data", 500)
		return
	}

}

// Starts server with all handler functions
//
// Returns an error if a problem with the server occurs
func (s *Server) Start() error {

	staticFS, err := fs.Sub(content, "static")
	if err != nil {
		return err
	}

	http.Handle("/", http.FileServer(http.FS(staticFS)))
	http.HandleFunc("/api", s.Api)
	http.HandleFunc("/api/schedule", s.ScheduleApi)
	http.HandleFunc("/api/status", s.StatusApi)
	http.HandleFunc("/api/uptime", s.UptimeAPI)
	http.HandleFunc("/api/errors", s.GetErrorData)
	http.HandleFunc("/api/db/size", s.GetDatabaseSize)
	http.HandleFunc("/api/db/clear", s.ClearDatabase)
	http.HandleFunc("/api/config", s.ReadConfigData)
	http.HandleFunc("/api/config/service", s.ConfigServiceApi)
	http.HandleFunc("/api/config/mqtt", s.ConfigMQTTApi)
	http.HandleFunc("/api/config/webhook", s.ConfigWebhookApi)
	log.Println("Starting server at http://localhost:8101/ . . .")
	if err := http.ListenAndServe(":8101", nil); err != nil {
		return err
	}

	return nil
}
