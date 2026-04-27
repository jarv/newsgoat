package config

import (
	"context"
	"strconv"

	"github.com/jarv/newsgoat/internal/database"
)

type Config struct {
	ReloadConcurrency   int
	ReloadTime          int  // Minutes between automatic reloads (0 = disabled)
	AutoReload          bool // Enable continuous automatic reloads
	SuppressFirstReload bool // Suppress the first automatic reload after startup
	ReloadOnStartup     bool // Reload all feeds on startup
	ThemeName           string
	HighlightStyle      string
	SpinnerType         string
	ShowReadFeeds       bool
	UnreadOnTop         bool // Show feeds with unread items at the top
}

// Setting keys
const (
	KeyReloadConcurrency   = "reload_concurrency"
	KeyReloadTime          = "reload_time"
	KeyAutoReload          = "auto_reload"
	KeySuppressFirstReload = "suppress_first_reload"
	KeyReloadOnStartup     = "reload_on_startup"
	KeyThemeName           = "theme_name"
	KeyHighlightStyle      = "highlight_style"
	KeySpinnerType         = "spinner_type"
	KeyShowReadFeeds       = "show_read_feeds"
	KeyUnreadOnTop         = "unread_on_top"
)

func GetDefaultConfig() Config {
	return Config{
		ReloadConcurrency:   4,
		ReloadTime:          60,    // 60 minutes default
		AutoReload:          true,  // Disabled by default
		SuppressFirstReload: false, // Don't suppress by default
		ReloadOnStartup:     true,  // Don't reload on startup by default
		ThemeName:           "dark",
		HighlightStyle:      "prefix-underline",
		SpinnerType:         "braille",
		ShowReadFeeds:       true,
		UnreadOnTop:         true, // Show unread feeds at top by default
	}
}

func LoadConfig(queries *database.Queries) (Config, error) {
	defaults := GetDefaultConfig()
	config := defaults

	// Load reload_concurrency
	if val, err := queries.GetSetting(context.Background(), KeyReloadConcurrency); err == nil {
		if intVal, err := strconv.Atoi(val.Value); err == nil {
			config.ReloadConcurrency = intVal
		}
	}

	// Load reload_time
	if val, err := queries.GetSetting(context.Background(), KeyReloadTime); err == nil {
		if intVal, err := strconv.Atoi(val.Value); err == nil {
			config.ReloadTime = intVal
		}
	}

	// Load auto_reload
	if val, err := queries.GetSetting(context.Background(), KeyAutoReload); err == nil {
		config.AutoReload = (val.Value == "true" || val.Value == "yes")
	}

	// Load suppress_first_reload
	if val, err := queries.GetSetting(context.Background(), KeySuppressFirstReload); err == nil {
		config.SuppressFirstReload = (val.Value == "true" || val.Value == "yes")
	}

	// Load reload_on_startup
	if val, err := queries.GetSetting(context.Background(), KeyReloadOnStartup); err == nil {
		config.ReloadOnStartup = (val.Value == "true" || val.Value == "yes")
	}

	// Load theme name
	if val, err := queries.GetSetting(context.Background(), KeyThemeName); err == nil {
		config.ThemeName = val.Value
	}

	// Load highlight style
	if val, err := queries.GetSetting(context.Background(), KeyHighlightStyle); err == nil {
		config.HighlightStyle = val.Value
	}

	// Load spinner type
	if val, err := queries.GetSetting(context.Background(), KeySpinnerType); err == nil {
		config.SpinnerType = val.Value
	}

	// Load show read feeds
	if val, err := queries.GetSetting(context.Background(), KeyShowReadFeeds); err == nil {
		config.ShowReadFeeds = (val.Value == "true" || val.Value == "yes")
	}

	// Load unread on top
	if val, err := queries.GetSetting(context.Background(), KeyUnreadOnTop); err == nil {
		config.UnreadOnTop = (val.Value == "true" || val.Value == "yes")
	}

	// Validate config values
	if config.ReloadConcurrency < 1 {
		config.ReloadConcurrency = 1
	}
	if config.ReloadConcurrency > 10 {
		config.ReloadConcurrency = 10
	}
	if config.ReloadTime < 0 {
		config.ReloadTime = 0
	}

	return config, nil
}

func SaveConfig(queries *database.Queries, config Config) error {
	// Save reload_concurrency
	if err := queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   KeyReloadConcurrency,
		Value: strconv.Itoa(config.ReloadConcurrency),
	}); err != nil {
		return err
	}

	// Save reload_time
	if err := queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   KeyReloadTime,
		Value: strconv.Itoa(config.ReloadTime),
	}); err != nil {
		return err
	}

	// Save auto_reload
	autoReloadStr := "false"
	if config.AutoReload {
		autoReloadStr = "true"
	}
	if err := queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   KeyAutoReload,
		Value: autoReloadStr,
	}); err != nil {
		return err
	}

	// Save suppress_first_reload
	suppressFirstReloadStr := "false"
	if config.SuppressFirstReload {
		suppressFirstReloadStr = "true"
	}
	if err := queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   KeySuppressFirstReload,
		Value: suppressFirstReloadStr,
	}); err != nil {
		return err
	}

	// Save reload_on_startup
	reloadOnStartupStr := "false"
	if config.ReloadOnStartup {
		reloadOnStartupStr = "true"
	}
	if err := queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   KeyReloadOnStartup,
		Value: reloadOnStartupStr,
	}); err != nil {
		return err
	}

	// Save theme name
	if err := queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   KeyThemeName,
		Value: config.ThemeName,
	}); err != nil {
		return err
	}

	// Save highlight style
	if err := queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   KeyHighlightStyle,
		Value: config.HighlightStyle,
	}); err != nil {
		return err
	}

	// Save spinner type
	if err := queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   KeySpinnerType,
		Value: config.SpinnerType,
	}); err != nil {
		return err
	}

	// Save show read feeds
	showReadFeedsStr := "false"
	if config.ShowReadFeeds {
		showReadFeedsStr = "true"
	}
	if err := queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   KeyShowReadFeeds,
		Value: showReadFeedsStr,
	}); err != nil {
		return err
	}

	// Save unread on top
	unreadOnTopStr := "false"
	if config.UnreadOnTop {
		unreadOnTopStr = "true"
	}
	if err := queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   KeyUnreadOnTop,
		Value: unreadOnTopStr,
	}); err != nil {
		return err
	}

	return nil
}
