package config

import (
	"context"

	"github.com/jarv/newsgoat/internal/database"
)

// LocalSettingsManager wraps database.Queries to implement SettingsManager
type LocalSettingsManager struct {
	queries *database.Queries
}

// NewLocalSettingsManager creates a new LocalSettingsManager
func NewLocalSettingsManager(queries *database.Queries) *LocalSettingsManager {
	return &LocalSettingsManager{queries: queries}
}

func (m *LocalSettingsManager) GetSetting(key string) (string, error) {
	setting, err := m.queries.GetSetting(context.Background(), key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (m *LocalSettingsManager) SetSetting(key, value string) error {
	return m.queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   key,
		Value: value,
	})
}
