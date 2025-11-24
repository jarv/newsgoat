package config

// SettingsManager is the interface for settings operations.
// Both the local Queries wrapper and the gRPC Client implement this interface.
type SettingsManager interface {
	GetSetting(key string) (string, error)
	SetSetting(key, value string) error
}
