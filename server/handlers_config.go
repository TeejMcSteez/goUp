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

// BackoffPayload is the request/response body for GET/POST /api/config/backoff.
type BackoffPayload struct {
	Backoff_Period string `json:"backoff_period"`
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
	slData := utils.ReadConfigSlack(utils.Current_Config)
	tgData := utils.ReadConfigTelegram(utils.Current_Config)
	haData := utils.ReadConfigHA(utils.Current_Config)
	dData := utils.ReadConfigDiscord(utils.Current_Config)
	backoffData := utils.ReadConfigBackoff(utils.Current_Config)

	data := utils.ConfigData{Services: sData, MQTT: mData, Webhook: wData, SMTP: eData, Gotify: gData, Slack: slData, Telegram: tgData, HA: haData, Discord: dData, Backoff_Period: backoffData}

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

// @Summary Add a monitored service
// @Tags config
// @Accept json
// @Produce json
// @Param service body utils.Service true "Service definition"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 409 {string} string "conflict"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/service [post]
func (s *Server) configServicePost(w http.ResponseWriter, req *http.Request) {
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
}

// @Summary Update a monitored service
// @Tags config
// @Accept json
// @Produce json
// @Param payload body ServiceUpdatePayload true "Old name + new service definition"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/service [put]
func (s *Server) configServicePut(w http.ResponseWriter, req *http.Request) {
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
}

// @Summary Delete a monitored service
// @Tags config
// @Accept json
// @Produce json
// @Param service body utils.Service true "Service definition"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/service [delete]
func (s *Server) configServiceDelete(w http.ResponseWriter, req *http.Request) {
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
}

func (s *Server) ConfigServiceApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		s.configServicePost(w, req)
	case "PUT":
		s.configServicePut(w, req)
	case "DELETE":
		s.configServiceDelete(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Set MQTT trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param mqtt body utils.MQTTTrigger true "MQTT config"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/mqtt [post]
func (s *Server) configMQTTPost(w http.ResponseWriter, req *http.Request) {
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
}

// @Summary Remove MQTT trigger configuration
// @Tags config
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 500 {string} string "internal server error"
// @Router /api/config/mqtt [delete]
func (s *Server) configMQTTDelete(w http.ResponseWriter, _ *http.Request) {
	if err := utils.DeleteConfigMQTT(utils.Current_Config); err != nil {
		log.Printf("Error deleting MQTT config: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write message", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigMQTTApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		s.configMQTTPost(w, req)
	case "DELETE":
		s.configMQTTDelete(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Set webhook trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param webhook body utils.WebhookTrigger true "Webhook config"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/webhook [post]
func (s *Server) configWebhookPost(w http.ResponseWriter, req *http.Request) {
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
}

// @Summary Remove webhook trigger configuration
// @Tags config
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 500 {string} string "internal server error"
// @Router /api/config/webhook [delete]
func (s *Server) configWebhookDelete(w http.ResponseWriter, _ *http.Request) {
	if err := utils.DeleteConfigWebhook(utils.Current_Config); err != nil {
		log.Printf("Error deleting webhook config: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write error message", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigWebhookApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		s.configWebhookPost(w, req)
	case "DELETE":
		s.configWebhookDelete(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Set SMTP trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param smtp body utils.SMTPTrigger true "SMTP config"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/smtp [post]
func (s *Server) configSMTPPost(w http.ResponseWriter, req *http.Request) {
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
}

// @Summary Remove SMTP trigger configuration
// @Tags config
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 500 {string} string "internal server error"
// @Router /api/config/smtp [delete]
func (s *Server) configSMTPDelete(w http.ResponseWriter, _ *http.Request) {
	if err := utils.DeleteConfigSMTPTrigger(utils.Current_Config); err != nil {
		log.Printf("Error deleting SMTP trigger: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write error message", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigSMTPApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		s.configSMTPPost(w, req)
	case "DELETE":
		s.configSMTPDelete(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Set Gotify trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param gotify body utils.GotifyTrigger true "Gotify config"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/gotify [post]
func (s *Server) configGotifyPost(w http.ResponseWriter, req *http.Request) {
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
}

// @Summary Remove Gotify trigger configuration
// @Tags config
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 500 {string} string "internal server error"
// @Router /api/config/gotify [delete]
func (s *Server) configGotifyDelete(w http.ResponseWriter, _ *http.Request) {
	if err := utils.DeleteConfigGotifyTrigger(utils.Current_Config); err != nil {
		log.Printf("Error deleting Gotify trigger: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write error message", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigGotifyApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		s.configGotifyPost(w, req)
	case "DELETE":
		s.configGotifyDelete(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Set Slack trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param slack body utils.SlackTrigger true "Slack config"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/slack [post]
func (s *Server) configSlackPost(w http.ResponseWriter, req *http.Request) {
	var slack utils.SlackTrigger
	if err := json.NewDecoder(req.Body).Decode(&slack); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.AddConfigSlackTrigger(utils.Current_Config, slack); err != nil {
		log.Printf("Error adding Slack trigger: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// @Summary Remove Slack trigger configuration
// @Tags config
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 500 {string} string "internal server error"
// @Router /api/config/slack [delete]
func (s *Server) configSlackDelete(w http.ResponseWriter, _ *http.Request) {
	if err := utils.DeleteConfigSlackTrigger(utils.Current_Config); err != nil {
		log.Printf("Error deleting Slack trigger: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigSlackApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		s.configSlackPost(w, req)
	case "DELETE":
		s.configSlackDelete(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Set Telegram trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param telegram body utils.TelegramTrigger true "Telegram config"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/telegram [post]
func (s *Server) configTelegramPost(w http.ResponseWriter, req *http.Request) {
	var telegram utils.TelegramTrigger
	if err := json.NewDecoder(req.Body).Decode(&telegram); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.AddConfigTelegramTrigger(utils.Current_Config, telegram); err != nil {
		log.Printf("Error adding Telegram trigger: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// @Summary Remove Telegram trigger configuration
// @Tags config
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 500 {string} string "internal server error"
// @Router /api/config/telegram [delete]
func (s *Server) configTelegramDelete(w http.ResponseWriter, _ *http.Request) {
	if err := utils.DeleteConfigTelegramTrigger(utils.Current_Config); err != nil {
		log.Printf("Error deleting Telegram trigger: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigTelegramApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		s.configTelegramPost(w, req)
	case "DELETE":
		s.configTelegramDelete(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Set Home Assistant trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param ha body utils.HATrigger true "Home Assistant config"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/ha [post]
func (s *Server) configHAPost(w http.ResponseWriter, req *http.Request) {
	var ha utils.HATrigger
	if err := json.NewDecoder(req.Body).Decode(&ha); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.AddConfigHATrigger(utils.Current_Config, ha); err != nil {
		log.Printf("Error adding Home Assistant trigger: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// @Summary Remove Home Assistant trigger configuration
// @Tags config
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 500 {string} string "internal server error"
// @Router /api/config/ha [delete]
func (s *Server) configHADelete(w http.ResponseWriter, _ *http.Request) {
	if err := utils.DeleteConfigHATrigger(utils.Current_Config); err != nil {
		log.Printf("Error deleting Home Assistant trigger: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigHAApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		s.configHAPost(w, req)
	case "DELETE":
		s.configHADelete(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Set Discord trigger configuration
// @Tags config
// @Accept json
// @Produce json
// @Param discord body utils.DiscordTrigger true "Discord config"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/discord [post]
func (s *Server) configDiscordPost(w http.ResponseWriter, req *http.Request) {
	var discord utils.DiscordTrigger
	if err := json.NewDecoder(req.Body).Decode(&discord); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.AddConfigDiscordTrigger(utils.Current_Config, discord); err != nil {
		log.Printf("Error adding Discord trigger: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// @Summary Remove Discord trigger configuration
// @Tags config
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 500 {string} string "internal server error"
// @Router /api/config/discord [delete]
func (s *Server) configDiscordDelete(w http.ResponseWriter, _ *http.Request) {
	if err := utils.DeleteConfigDiscordTrigger(utils.Current_Config); err != nil {
		log.Printf("Error deleting Discord trigger: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprintf(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigDiscordApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "POST":
		s.configDiscordPost(w, req)
	case "DELETE":
		s.configDiscordDelete(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Get the global trigger backoff period
// @Tags config
// @Produce json
// @Success 200 {object} BackoffPayload
// @Failure 500 {string} string "internal server error"
// @Router /api/config/backoff [get]
func (s *Server) configBackoffGet(w http.ResponseWriter, _ *http.Request) {
	payload := BackoffPayload{Backoff_Period: utils.ReadConfigBackoff(utils.Current_Config)}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Set the global trigger backoff period
// @Tags config
// @Accept json
// @Produce json
// @Param backoff body BackoffPayload true "Backoff duration, e.g. 5m; blank disables the global backoff"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/backoff [post]
func (s *Server) configBackoffPost(w http.ResponseWriter, req *http.Request) {
	var payload BackoffPayload
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.UpdateBackoff(utils.Current_Config, payload.Backoff_Period, "global"); err != nil {
		log.Printf("Error updating global backoff: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigBackoffApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "GET":
		s.configBackoffGet(w, req)
	case "POST":
		s.configBackoffPost(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}

// @Summary Get a service's active (fetch enabled) state
// @Tags config
// @Produce json
// @Param name query string true "Service name to look up"
// @Success 200 {object} utils.Service
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/service/active [get]
func (s *Server) configActiveGet(w http.ResponseWriter, req *http.Request) {
	// This is not really used in client as API returns the Service which contains active state passed to component
	// However, this is mainly for the API to explicitly get active state for a service
	name := req.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing required query parameter: name", http.StatusBadRequest)
		return
	}
	svc, err := utils.ReadConfigService(utils.Current_Config, utils.Service{Name: name})
	if err != nil {
		http.Error(w, fmt.Sprintf("could not read service: %v", err), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(svc); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// @Summary Enable or disable fetch for a service
// @Tags config
// @Accept json
// @Produce json
// @Param service body utils.Service true "Service to toggle active state"
// @Success 200 {object} map[string]bool
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/config/service/active [post]
func (s *Server) configActivePost(w http.ResponseWriter, req *http.Request) {
	var service utils.Service
	if err := json.NewDecoder(req.Body).Decode(&service); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.UpdateConfigServiceActive(utils.Current_Config, service); err != nil {
		http.Error(w, fmt.Sprintf("could not update service: %v", err), http.StatusInternalServerError)
		return
	}
	if err := utils.Setup(utils.Current_Config); err != nil {
		log.Printf("Warning: failed to refresh endpoints after toggling service active state: %v", err)
	}
	if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (s *Server) ConfigActiveApi(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	switch req.Method {
	case "GET":
		s.configActiveGet(w, req)
	case "POST":
		s.configActivePost(w, req)
	default:
		http.Error(w, "Invalid method", http.StatusBadRequest)
	}
}
