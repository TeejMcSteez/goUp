package server

import (
	"encoding/json"
	"fmt"
	"goUp/utils"
	"log"
	"net/http"
)

// ServiceUpdatePayload is the request body for PUT /api/config/service.
type ServiceUpdatePayload struct {
	OldName string        `json:"old_name"`
	Service utils.Service `json:"service"`
}

// @Summary Get full configuration
// @Tags config
// @Produce json
// @Success 200 {object} utils.ConfigData
// @Failure 500 {string} string "internal server error"
// @Router /api/config [get]
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

// @Summary Add, update, or delete a monitored service
// @Tags config
// @Accept json
// @Produce json
// @Param service body utils.Service false "Service definition (POST and DELETE)"
// @Param payload body ServiceUpdatePayload false "Old name + new service definition (PUT)"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 409 {string} string "conflict"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/service [post]
// @Router /api/config/service [put]
// @Router /api/config/service [delete]
func (s *Server) ConfigServiceApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		var service utils.Service
		if err := json.NewDecoder(req.Body).Decode(&service); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, valid := utils.Current_Config.Services[service.URL]
		if !valid {
			http.Error(w, "URL for this service is already in the configuration", http.StatusConflict)
			return
		}
		if err := utils.AddConfigService(utils.Current_Config, service); err != nil {
			log.Printf("Error adding config service: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := utils.Setup(utils.Current_Config); err != nil {
			log.Printf("Warning: failed to refresh endpoints after adding service: %v", err)
		}
		if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write message", http.StatusInternalServerError)
		}
	case "PUT":
		var payload ServiceUpdatePayload
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := utils.UpdateConfigService(utils.Current_Config, payload.OldName, payload.Service, s.db); err != nil {
			log.Printf("Error updating config service: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := utils.Setup(utils.Current_Config); err != nil {
			log.Printf("Warning: failed to refresh endpoints after updating service: %v", err)
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
		if err := utils.DeleteConfigService(utils.Current_Config, service, s.db); err != nil {
			log.Printf("Error deleting config service: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := utils.Setup(utils.Current_Config); err != nil {
			log.Printf("Warning: failed to refresh endpoints after deleting service: %v", err)
		}
		if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write message", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Set or remove MQTT trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param mqtt body utils.MQTTTrigger false "MQTT config (POST only)"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/mqtt [post]
// @Router /api/config/mqtt [delete]
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

// @Summary Set or remove webhook trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param webhook body utils.WebhookTrigger false "Webhook config (POST only)"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/webhook [post]
// @Router /api/config/webhook [delete]
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
