package server

import (
	"encoding/json"
	"fmt"
	"goUp/utils"
	"log"
	"net/http"
)

// Reads all configuration data currently in memory
func (s *Server) ReadConfigData(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sData := utils.ReadConfigServices(utils.Current_Config)
	mData := utils.ReadConfigMQTT(utils.Current_Config)
	wData := utils.ReadConfigWebhook(utils.Current_Config)

	data := utils.ConfigData{Services: sData, MQTT: mData, Webhook: wData}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding configuration data to json: %v", err)
		http.Error(w, "Server Error: Failed to parse config data", 500)
		return
	}
}

// Handles all methods for services in configuration
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
		if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write message", http.StatusInternalServerError)
		}
	case "PUT":
		var payload struct {
			OldName string        `json:"old_name"`
			Service utils.Service `json:"service"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := utils.UpdateConfigService(utils.Current_Config, payload.OldName, payload.Service); err != nil {
			log.Printf("Error updating config service: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write message", http.StatusInternalServerError)
		}
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
		if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write message", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// Handles all methods for MQTT trigger in configuration
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
		if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write message", http.StatusInternalServerError)
		}
	case "DELETE":
		if err := utils.DeleteConfigMQTT(utils.Current_Config); err != nil {
			log.Printf("Error deleting MQTT config: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write message", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// Handles all methods for Webhook Trigger in configuration
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
		if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write error message", http.StatusInternalServerError)
		}
	case "DELETE":
		if err := utils.DeleteConfigTrigger(utils.Current_Config); err != nil {
			log.Printf("Error deleting webhook config: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write error message", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}
