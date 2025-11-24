package config

import (
	"strconv"
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
	CheckForUpdates     bool // Check for updates on launch
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
	KeyCheckForUpdates     = "check_for_updates"
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
		CheckForUpdates:     true, // Check for updates on launch by default
	}
}

func LoadConfig(settings SettingsManager) (Config, error) {
	defaults := GetDefaultConfig()
	config := defaults

	// Load reload_concurrency
	if val, err := settings.GetSetting(KeyReloadConcurrency); err == nil {
		if intVal, err := strconv.Atoi(val); err == nil {
			config.ReloadConcurrency = intVal
		}
	}

	// Load reload_time
	if val, err := settings.GetSetting(KeyReloadTime); err == nil {
		if intVal, err := strconv.Atoi(val); err == nil {
			config.ReloadTime = intVal
		}
	}

	// Load auto_reload
	if val, err := settings.GetSetting(KeyAutoReload); err == nil {
		config.AutoReload = (val == "true" || val == "yes")
	}

	// Load suppress_first_reload
	if val, err := settings.GetSetting(KeySuppressFirstReload); err == nil {
		config.SuppressFirstReload = (val == "true" || val == "yes")
	}

	// Load reload_on_startup
	if val, err := settings.GetSetting(KeyReloadOnStartup); err == nil {
		config.ReloadOnStartup = (val == "true" || val == "yes")
	}

	// Load theme name
	if val, err := settings.GetSetting(KeyThemeName); err == nil {
		config.ThemeName = val
	}

	// Load highlight style
	if val, err := settings.GetSetting(KeyHighlightStyle); err == nil {
		config.HighlightStyle = val
	}

	// Load spinner type
	if val, err := settings.GetSetting(KeySpinnerType); err == nil {
		config.SpinnerType = val
	}

	// Load show read feeds
	if val, err := settings.GetSetting(KeyShowReadFeeds); err == nil {
		config.ShowReadFeeds = (val == "true" || val == "yes")
	}

	// Load unread on top
	if val, err := settings.GetSetting(KeyUnreadOnTop); err == nil {
		config.UnreadOnTop = (val == "true" || val == "yes")
	}

	// Load check for updates
	if val, err := settings.GetSetting(KeyCheckForUpdates); err == nil {
		config.CheckForUpdates = (val == "true" || val == "yes")
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

func SaveConfig(settings SettingsManager, config Config) error {
	// Save reload_concurrency
	if err := settings.SetSetting(KeyReloadConcurrency, strconv.Itoa(config.ReloadConcurrency)); err != nil {
		return err
	}

	// Save reload_time
	if err := settings.SetSetting(KeyReloadTime, strconv.Itoa(config.ReloadTime)); err != nil {
		return err
	}

	// Save auto_reload
	autoReloadStr := "false"
	if config.AutoReload {
		autoReloadStr = "true"
	}
	if err := settings.SetSetting(KeyAutoReload, autoReloadStr); err != nil {
		return err
	}

	// Save suppress_first_reload
	suppressFirstReloadStr := "false"
	if config.SuppressFirstReload {
		suppressFirstReloadStr = "true"
	}
	if err := settings.SetSetting(KeySuppressFirstReload, suppressFirstReloadStr); err != nil {
		return err
	}

	// Save reload_on_startup
	reloadOnStartupStr := "false"
	if config.ReloadOnStartup {
		reloadOnStartupStr = "true"
	}
	if err := settings.SetSetting(KeyReloadOnStartup, reloadOnStartupStr); err != nil {
		return err
	}

	// Save theme name
	if err := settings.SetSetting(KeyThemeName, config.ThemeName); err != nil {
		return err
	}

	// Save highlight style
	if err := settings.SetSetting(KeyHighlightStyle, config.HighlightStyle); err != nil {
		return err
	}

	// Save spinner type
	if err := settings.SetSetting(KeySpinnerType, config.SpinnerType); err != nil {
		return err
	}

	// Save show read feeds
	showReadFeedsStr := "false"
	if config.ShowReadFeeds {
		showReadFeedsStr = "true"
	}
	if err := settings.SetSetting(KeyShowReadFeeds, showReadFeedsStr); err != nil {
		return err
	}

	// Save unread on top
	unreadOnTopStr := "false"
	if config.UnreadOnTop {
		unreadOnTopStr = "true"
	}
	if err := settings.SetSetting(KeyUnreadOnTop, unreadOnTopStr); err != nil {
		return err
	}

	// Save check for updates
	checkForUpdatesStr := "false"
	if config.CheckForUpdates {
		checkForUpdatesStr = "true"
	}
	if err := settings.SetSetting(KeyCheckForUpdates, checkForUpdatesStr); err != nil {
		return err
	}

	return nil
}
