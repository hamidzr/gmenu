package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"

	"github.com/spf13/viper"
)

type configKeyVariant struct {
	canonical string
	camel     string
}

var configKeyVariants = []configKeyVariant{
	{canonical: "title"},
	{canonical: "prompt"},
	{canonical: "menu_id", camel: "menuId"},
	{canonical: "search_method", camel: "searchMethod"},
	{canonical: "preserve_order", camel: "preserveOrder"},
	{canonical: "initial_query", camel: "initialQuery"},
	{canonical: "auto_accept", camel: "autoAccept"},
	{canonical: "terminal_mode", camel: "terminalMode"},
	{canonical: "no_numeric_selection", camel: "noNumericSelection"},
	{canonical: "min_width", camel: "minWidth"},
	{canonical: "min_height", camel: "minHeight"},
	{canonical: "max_width", camel: "maxWidth"},
	{canonical: "max_height", camel: "maxHeight"},
	{canonical: "accept_custom_selection", camel: "acceptCustomSelection"},
}

var canonicalByKey = func() map[string]string {
	m := make(map[string]string, len(configKeyVariants)*2)
	for _, variant := range configKeyVariants {
		m[variant.canonical] = variant.canonical
		if variant.camel != "" {
			m[variant.camel] = variant.canonical
		}
	}
	return m
}()

func registerConfigKeyAliases(v *viper.Viper) {
	for _, variant := range configKeyVariants {
		if variant.camel != "" {
			v.RegisterAlias(variant.camel, variant.canonical)
		}
	}
}

func validateConfigFileKeys(configPath string) error {
	if configPath == "" {
		return nil
	}

	displayPath := configFileDisplayPath(configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file %s: %w", displayPath, err)
	}

	if len(data) == 0 {
		return nil
	}

	var raw map[string]interface{}
	if err := yamlv3.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error parsing config file %s: %w", displayPath, err)
	}

	seen := make(map[string]string, len(raw))
	for key := range raw {
		canonical, ok := canonicalByKey[key]
		if !ok {
			return fmt.Errorf("config file %s contains invalid key %q", displayPath, key)
		}
		if previous, exists := seen[canonical]; exists && previous != key {
			return fmt.Errorf("config file %s contains both %q (%s) and %q (%s); use one naming style for %q", displayPath, previous, keyStyle(previous), key, keyStyle(key), canonical)
		}
		seen[canonical] = key
	}

	return nil
}

func keyStyle(key string) string {
	if strings.Contains(key, "_") {
		return "snake_case"
	}
	if strings.Contains(key, "-") {
		return "kebab-case"
	}
	if len(key) == 0 {
		return "unknown style"
	}
	return "camelCase"
}

func configFileDisplayPath(path string) string {
	if path == "" {
		return path
	}
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}
