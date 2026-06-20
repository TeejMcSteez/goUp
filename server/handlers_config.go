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

// Used as the response for GET request
type DatabaseSizePayload struct {
	Size string `json:"db_max_size"`
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
	eData := utils.ReadConfigSMTP(utils.Current_Config)
	gData := utils.ReadConfigGotify(utils.Current_Config)

	data := utils.ConfigData{Services: sData, MQTT: mData, Webhook: wData, SMTP: eData, Gotify: gData}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding configuration data to json: %v", err)
		http.Error(w, "Server Error: Failed to parse config data", 500)
		return
	}
}

// @Summary Get the configured database max size
// @Tags config
// @Produce json
// @Success 200 {object} DatabaseSizePayload
// @Failure 500 {string} string "internal server error"
// @Router /api/config/size [get]
func (s *Server) configDatabaseGet(w http.ResponseWriter, _ *http.Request) {
	size, err := utils.ReadDatabaseSize(utils.Current_Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(DatabaseSizePayload{Size: size}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Update the database max size
// @Tags config
// @Accept json
// @Produce json
// @Param db_max_size body DatabaseSizePayload true "Size string (e.g. 1kb, 2mb, 3gb)"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/size [post]
func (s *Server) configDatabasePost(w http.ResponseWriter, req *http.Request) {
	var data DatabaseSizePayload
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := utils.UpdateDatabaseSize(utils.Current_Config, data.Size); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write message", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigDatabaseApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "GET":
		s.configDatabaseGet(w, req)
	case "POST":
		s.configDatabasePost(w, req)
	default:
		http.Error(w, "Bad method", http.StatusBadRequest)
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
		for _, svc := range utils.Current_Config.Services {
			if svc.URL == service.URL {
				http.Error(w, "URL for this service is already in the configuration", http.StatusConflict)
				return
			}
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
		if err := utils.DbGarbageCollect(s.db, utils.Current_Config); err != nil {
			log.Printf("Warning: GC failed after updating service: %v", err)
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
		if err := utils.DbGarbageCollect(s.db, utils.Current_Config); err != nil {
			log.Printf("Warning: GC failed after deleting service: %v", err)
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

// @Summary Set or remove SMTP trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param smtp body utils.SMTPTrigger false "SMTP config (POST only)"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/smtp [post]
// @Router /api/config/smtp [delete]
func (s *Server) ConfigSMTPApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		var smtp utils.SMTPTrigger
		if err := json.NewDecoder(req.Body).Decode(&smtp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := utils.AddConfigSMTPTrigger(utils.Current_Config, smtp); err != nil {
			log.Printf("Error adding SMTP trigger: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write error message", http.StatusInternalServerError)
		}
	case "DELETE":
		if err := utils.DeleteConfigSMTPTrigger(utils.Current_Config); err != nil {
			log.Printf("Error deleting SMTP trigger: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write error message", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Set or remove Gotify trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param gotify body utils.GotifyTrigger false "Gotify config (POST only)"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/gotify [post]
// @Router /api/config/gotify [delete]
func (s *Server) ConfigGotifyApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		var gotify utils.GotifyTrigger
		if err := json.NewDecoder(req.Body).Decode(&gotify); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := utils.AddConfigGotifyTrigger(utils.Current_Config, gotify); err != nil {
			log.Printf("Error adding Gotify trigger: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write error message", http.StatusInternalServerError)
		}
	case "DELETE":
		if err := utils.DeleteConfigGotifyTrigger(utils.Current_Config); err != nil {
			log.Printf("Error deleting Gotify trigger: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
			http.Error(w, "Failed to write error message", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}
