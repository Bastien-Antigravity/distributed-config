package models

// Telebot Capability
// -----------------------------------------------------------------------------

type TelebotCapability struct {
	Token  string `yaml:"token"`
	ChatID string `yaml:"chat_id"`
	IP     string `yaml:"ip"`
	Port   string `yaml:"port"`
}
