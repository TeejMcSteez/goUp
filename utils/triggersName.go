package utils

func (m *MQTTTrigger) Name() string     { return "MQTT" }
func (w *WebhookTrigger) Name() string  { return "Webhook" }
func (s *SMTPTrigger) Name() string     { return "SMTP" }
func (g *GotifyTrigger) Name() string   { return "Gotify" }
func (s *SlackTrigger) Name() string    { return "Slack" }
func (t *TelegramTrigger) Name() string { return "Telegram" }
func (h *HATrigger) Name() string       { return "Home Assistant" }
func (d *DiscordTrigger) Name() string  { return "Discord" }
