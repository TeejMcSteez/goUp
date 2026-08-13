package utils

func (m *MQTTTrigger) IsConfigured() bool {
	return m.Mqtt_broker != nil
}

// IsConfigured reports whether the webhook trigger has a URL set.
func (w *WebhookTrigger) IsConfigured() bool {
	return w.Webhook_url != nil
}

// Checks if all parameters for the SMTP trigger is is configured
func (e *SMTPTrigger) IsConfigured() bool {
	return e.SMTPServer != nil && e.Email != nil && e.App_Password != nil
}

func (g *GotifyTrigger) IsConfigured() bool {
	return g.Gotify_Server != nil && g.Gotify_Token != nil
}

func (s *SlackTrigger) IsConfigured() bool {
	return s.Slack_Token != nil && s.Slack_Channel != nil
}

func (t *TelegramTrigger) IsConfigured() bool {
	return t.Telegram_Channel_Id != nil && t.Telegram_Token != nil
}

func (h *HATrigger) IsConfigured() bool {
	return h.HA_URL != nil && h.HA_Token != nil
}

func (d *DiscordTrigger) IsConfigured() bool {
	return d.Discord_Auth != nil && d.Discord_Channel != nil
}
