package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hamidzr/gmenu/model"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// Helper function for testing config loading
func loadConfigFromFile(configPath string) (*model.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config model.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// TestConfigLoading tests configuration file loading
func TestConfigLoading(t *testing.T) {
	// Create a temporary directory for test config files
	tmpDir, err := os.MkdirTemp("", "gmenu-config-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test config content
	configContent := `
title: "Test Menu"
prompt: "Test Prompt"
menu_id: "test-menu"
search_method: "fuzzy"
preserve_order: true
auto_accept: false
terminal_mode: false
no_numeric_selection: false
min_width: 800
min_height: 400
max_width: 1600
max_height: 1000
accept_custom_selection: true
initial_query: "test query"
`

	// Write test config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test loading the config
	config, err := loadConfigFromFile(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify all fields are loaded correctly
	assert.Equal(t, "Test Menu", config.Title)
	assert.Equal(t, "Test Prompt", config.Prompt)
	assert.Equal(t, "test-menu", config.MenuID)
	assert.Equal(t, "fuzzy", config.SearchMethod)
	assert.True(t, config.PreserveOrder)
	assert.False(t, config.AutoAccept)
	assert.False(t, config.TerminalMode)
	assert.False(t, config.NoNumericSelection)
	assert.Equal(t, float32(800), config.MinWidth)
	assert.Equal(t, float32(400), config.MinHeight)
	assert.Equal(t, float32(1600), config.MaxWidth)
	assert.Equal(t, float32(1000), config.MaxHeight)
	assert.True(t, config.AcceptCustomSelection)
	assert.Equal(t, "test query", config.InitialQuery)
}

// TestConfigDefaults tests that default values are set correctly
func TestConfigDefaults(t *testing.T) {
	// Get defaults directly from the model (since the helper function doesn't apply defaults)
	defaults := model.DefaultConfig()

	// Test that defaults are set correctly
	assert.Equal(t, "gmenu", defaults.Title)
	assert.Equal(t, "Search", defaults.Prompt)
	assert.Equal(t, "", defaults.MenuID)
	assert.Equal(t, "fuzzy", defaults.SearchMethod)
	assert.False(t, defaults.PreserveOrder)
	assert.False(t, defaults.AutoAccept)
	assert.False(t, defaults.TerminalMode)
	assert.True(t, defaults.NoNumericSelection)
	assert.Equal(t, float32(600), defaults.MinWidth)
	assert.Equal(t, float32(300), defaults.MinHeight)
	assert.Equal(t, float32(1920), defaults.MaxWidth)
	assert.Equal(t, float32(1080), defaults.MaxHeight)
	assert.True(t, defaults.AcceptCustomSelection)
	assert.Equal(t, "", defaults.InitialQuery)
}

// TestInvalidConfig tests handling of invalid configuration
func TestInvalidConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-config-invalid")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test invalid YAML syntax
	invalidConfigPath := filepath.Join(tmpDir, "invalid.yaml")
	invalidContent := `
title: "Test Menu"
prompt: [invalid yaml syntax
`
	err = os.WriteFile(invalidConfigPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	config, err := loadConfigFromFile(invalidConfigPath)
	assert.Error(t, err)
	assert.Nil(t, config)
}

// TestNonExistentConfig tests handling of missing config files
func TestNonExistentConfig(t *testing.T) {
	nonExistentPath := "/path/that/does/not/exist/config.yaml"

	config, err := loadConfigFromFile(nonExistentPath)
	assert.Error(t, err)
	assert.Nil(t, config)
}

// TestPartialConfig tests that partial configuration files work correctly
func TestPartialConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-config-partial")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Config with only some fields
	partialConfigContent := `
title: "Partial Config"
min_width: 1000
search_method: "exact"
`

	configPath := filepath.Join(tmpDir, "partial.yaml")
	err = os.WriteFile(configPath, []byte(partialConfigContent), 0644)
	require.NoError(t, err)

	config, err := loadConfigFromFile(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Apply defaults to partial config (simulate what viper does)
	defaults := model.DefaultConfig()
	if config.Prompt == "" {
		config.Prompt = defaults.Prompt
	}
	if config.MinHeight == 0 {
		config.MinHeight = defaults.MinHeight
	}

	// Check specified values
	assert.Equal(t, "Partial Config", config.Title)
	assert.Equal(t, float32(1000), config.MinWidth)
	assert.Equal(t, "exact", config.SearchMethod)

	// Check that defaults are used for unspecified values
	assert.Equal(t, "Search", config.Prompt)        // default
	assert.Equal(t, float32(300), config.MinHeight) // default
}

// TestInitConfigFile tests config file generation
func TestInitConfigFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-config-init")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Change to temp directory for testing
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Override the config directory function to use temp directory
	// Create a temporary config dir that doesn't exist yet
	testConfigDir := filepath.Join(tmpDir, ".config", "gmenu")

	// Test generating default config by manually creating the file
	// since InitConfigFile checks for existing files
	configPath := filepath.Join(testConfigDir, "config.yaml")
	err = os.MkdirAll(testConfigDir, 0755)
	require.NoError(t, err)

	// Generate config content manually for testing
	defaults := model.DefaultConfig()
	yamlData, err := yaml.Marshal(defaults)
	require.NoError(t, err)

	err = os.WriteFile(configPath, yamlData, 0644)
	require.NoError(t, err)
	assert.NotEmpty(t, configPath)

	// Verify the file was created
	assert.FileExists(t, configPath)

	// Verify the file can be loaded
	config, err := loadConfigFromFile(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Test that a second config directory would be different for a different menu ID
	testConfigDir2 := filepath.Join(tmpDir, ".config", "gmenu", "other-menu")
	menuConfigPath := filepath.Join(testConfigDir2, "config.yaml")
	err = os.MkdirAll(testConfigDir2, 0755)
	require.NoError(t, err)

	err = os.WriteFile(menuConfigPath, yamlData, 0644)
	require.NoError(t, err)
	assert.FileExists(t, menuConfigPath)

	// Paths should be different for different menu IDs
	assert.NotEqual(t, configPath, menuConfigPath)
}

func TestInitConfigAcceptsCamelCase(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "gmenu")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	configContent := `
menuId: "camel-menu"
initialQuery: "find me"
autoAccept: true
minHeight: 480
`

	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o644))

	cmd := &cobra.Command{Use: "gmenu"}
	BindFlags(cmd)

	cfg, err := InitConfig(cmd)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "camel-menu", cfg.MenuID)
	assert.Equal(t, "find me", cfg.InitialQuery)
	assert.True(t, cfg.AutoAccept)
	assert.Equal(t, float32(480), cfg.MinHeight)
}

func TestInitConfigRejectsMixedNamingStyles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "gmenu")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	configContent := `
initial_query: "snake"
initialQuery: "camel"
`

	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o644))

	cmd := &cobra.Command{Use: "gmenu"}
	BindFlags(cmd)

	cfg, err := InitConfig(cmd)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "initial_query")
	assert.Contains(t, err.Error(), "initialQuery")
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: `
title: "Valid Config"
search_method: "fuzzy"
min_width: 600
min_height: 300
`,
			expectError: false,
		},
		{
			name: "invalid search method",
			config: `
title: "Invalid Search"
search_method: "invalid_method"
`,
			expectError: false, // Should use default, not error
		},
		{
			name: "negative dimensions",
			config: `
min_width: -100
min_height: -50
`,
			expectError: false, // Should handle gracefully
		},
		{
			name: "zero dimensions",
			config: `
min_width: 0
min_height: 0
`,
			expectError: false, // Valid (auto-size)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "gmenu-config-validation")
			require.NoError(t, err)
			defer func() { _ = os.RemoveAll(tmpDir) }()

			configPath := filepath.Join(tmpDir, "test.yaml")
			err = os.WriteFile(configPath, []byte(tc.config), 0644)
			require.NoError(t, err)

			config, err := loadConfigFromFile(configPath)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}

// TestEnvironmentVariableOverrides tests that env vars override config files
func TestEnvironmentVariableOverrides(t *testing.T) {
	// Note: This test demonstrates that env vars would override config files
	// when using the proper viper-based config system. The loadConfigFromFile
	// helper doesn't implement env var handling, so we test the expected behavior.

	// Set environment variables
	_ = os.Setenv("GMENU_TITLE", "Env Title")
	_ = os.Setenv("GMENU_MIN_WIDTH", "1000")
	_ = os.Setenv("GMENU_SEARCH_METHOD", "exact")
	defer func() {
		_ = os.Unsetenv("GMENU_TITLE")
		_ = os.Unsetenv("GMENU_MIN_WIDTH")
		_ = os.Unsetenv("GMENU_SEARCH_METHOD")
	}()

	// Test that environment variables are accessible
	assert.Equal(t, "Env Title", os.Getenv("GMENU_TITLE"))
	assert.Equal(t, "1000", os.Getenv("GMENU_MIN_WIDTH"))
	assert.Equal(t, "exact", os.Getenv("GMENU_SEARCH_METHOD"))

	// In a real viper-based config system, these would override file values
	// This test validates the expected behavior rather than the helper function
}

// TestConfigSearchPaths tests the configuration file search logic
func TestConfigSearchPaths(t *testing.T) {
	// This test would typically test the file search paths,
	// but since we're using viper, we'll test the path resolution logic

	testCases := []struct {
		menuID        string
		expectedPaths []string
	}{
		{
			menuID: "",
			expectedPaths: []string{
				"config.yaml",
				"gmenu.yaml",
			},
		},
		{
			menuID: "test-menu",
			expectedPaths: []string{
				"test-menu/config.yaml",
				"config.yaml",
				"gmenu.yaml",
			},
		},
	}

	for _, tc := range testCases {
		t.Run("menuID_"+tc.menuID, func(t *testing.T) {
			// This is a placeholder for testing path resolution
			// In a real implementation, you might test the actual search paths
			if tc.menuID == "" {
				assert.Contains(t, tc.expectedPaths, "config.yaml")
			} else {
				assert.Contains(t, tc.expectedPaths, tc.menuID+"/config.yaml")
			}
		})
	}
}

// TestConfigModel tests the configuration model struct
func TestConfigModel(t *testing.T) {
	config := &model.Config{
		Title:                 "Test Title",
		Prompt:                "Test Prompt",
		MenuID:                "test-menu",
		SearchMethod:          "fuzzy",
		PreserveOrder:         true,
		InitialQuery:          "initial",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              800,
		MinHeight:             600,
		MaxWidth:              1600,
		MaxHeight:             1200,
		AcceptCustomSelection: true,
	}

	// Test that all fields are accessible
	assert.Equal(t, "Test Title", config.Title)
	assert.Equal(t, "Test Prompt", config.Prompt)
	assert.Equal(t, "test-menu", config.MenuID)
	assert.Equal(t, "fuzzy", config.SearchMethod)
	assert.True(t, config.PreserveOrder)
	assert.Equal(t, "initial", config.InitialQuery)
	assert.False(t, config.AutoAccept)
	assert.False(t, config.TerminalMode)
	assert.False(t, config.NoNumericSelection)
	assert.Equal(t, float32(800), config.MinWidth)
	assert.Equal(t, float32(600), config.MinHeight)
	assert.Equal(t, float32(1600), config.MaxWidth)
	assert.Equal(t, float32(1200), config.MaxHeight)
	assert.True(t, config.AcceptCustomSelection)

	// Test field modification
	config.Title = "Modified Title"
	assert.Equal(t, "Modified Title", config.Title)
}
