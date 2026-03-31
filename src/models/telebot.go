package models

// Telebot Capability
// -----------------------------------------------------------------------------

type TelebotCapability struct {
	Token  string `yaml:"token" json:"token"`
	ChatID string `yaml:"chat_id" json:"chat_id"`
	IP     string `yaml:"ip" json:"ip"`
	Port   string `yaml:"port" json:"port"`
}
