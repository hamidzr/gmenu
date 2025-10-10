package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hamidzr/gmenu/model"
	"gopkg.in/yaml.v2"
)

const comboSwitcherMenuID = "combo-switcher"

var comboSwitcherCanonicalKeys = map[string]string{
	"title":                 "title",
	"prompt":                "prompt",
	"menuid":                "menu_id",
	"searchmethod":          "search_method",
	"preserveorder":         "preserve_order",
	"initialquery":          "initial_query",
	"autoaccept":            "auto_accept",
	"terminalmode":          "terminal_mode",
	"nonumericselection":    "no_numeric_selection",
	"minwidth":              "min_width",
	"minheight":             "min_height",
	"maxwidth":              "max_width",
	"maxheight":             "max_height",
	"acceptcustomselection": "accept_custom_selection",
}

// getConfigPaths returns the config directory paths in priority order
// prefers ~/.config over macos application support dir
func GetConfigPaths(menuID string) []string {
	var paths []string

	// when menu ID is provided, prioritize namespaced configs
	if menuID != "" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			paths = append(paths, filepath.Join(homeDir, ".config", "gmenu", menuID))
			paths = append(paths, filepath.Join(homeDir, ".gmenu", menuID))
		}
		if configDir, err := os.UserConfigDir(); err == nil {
			paths = append(paths, filepath.Join(configDir, "gmenu", menuID))
		}
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".config", "gmenu"))
		paths = append(paths, filepath.Join(homeDir, ".gmenu"))
	}
	if configDir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(configDir, "gmenu"))
	}

	return paths
}

// getPreferredConfigDir returns the preferred config directory for writing
func GetPreferredConfigDir(menuID string) (string, error) {
	if homeDir, err := os.UserHomeDir(); err == nil {
		if menuID != "" {
			return filepath.Join(homeDir, ".config", "gmenu", menuID), nil
		}
		return filepath.Join(homeDir, ".config", "gmenu"), nil
	}

	if userConfigDir, err := os.UserConfigDir(); err == nil {
		if menuID != "" {
			return filepath.Join(userConfigDir, "gmenu", menuID), nil
		}
		return filepath.Join(userConfigDir, "gmenu"), nil
	}

	return "", fmt.Errorf("unable to determine config directory")
}

// GetConfigByMenuID loads the config for a given menu ID from the filesystem.
// It searches for the first config.yaml file in the config paths for the menu ID.
func GetConfigByMenuID(menuID string) (*model.Config, error) {
	paths := GetConfigPaths(menuID)
	for _, dir := range paths {
		configPath := filepath.Join(dir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, err
			}

			if menuID == comboSwitcherMenuID {
				if data, err = normalizeComboSwitcherConfig(data); err != nil {
					return nil, err
				}
			}

			cfg := model.DefaultConfig()
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, err
			}
			return cfg, nil
		}
	}
	return nil, fmt.Errorf("no config.yaml found for menu id '%s'", menuID)
}

func normalizeComboSwitcherConfig(data []byte) ([]byte, error) {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return data, nil
	}

	normalized := make(map[string]interface{}, len(raw))
	seen := make(map[string]string, len(raw))

	for key, value := range raw {
		normalizedKey := normalizeKeyVariant(key)
		canonical, ok := comboSwitcherCanonicalKeys[normalizedKey]
		if !ok {
			canonical = key
		}

		if previous, exists := seen[canonical]; exists && previous != key {
			return nil, fmt.Errorf("duplicate config keys %q and %q resolve to %q", previous, key, canonical)
		}

		seen[canonical] = key
		normalized[canonical] = convertNestedMaps(value)
	}

	return yaml.Marshal(normalized)
}

func normalizeKeyVariant(key string) string {
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, "-", "")
	key = strings.ReplaceAll(key, " ", "")
	return key
}

func convertNestedMaps(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		converted := make(map[string]interface{}, len(v))
		for key, nested := range v {
			converted[key] = convertNestedMaps(nested)
		}
		return converted
	case map[interface{}]interface{}:
		converted := make(map[string]interface{}, len(v))
		for key, nested := range v {
			converted[fmt.Sprint(key)] = convertNestedMaps(nested)
		}
		return converted
	case []interface{}:
		for i := range v {
			v[i] = convertNestedMaps(v[i])
		}
		return v
	default:
		return value
	}
}
