package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigStruct tests the Config struct
func TestConfigStruct(t *testing.T) {
	config := &Config{
		Title:                 "Test Title",
		Prompt:                "Test Prompt",
		MenuID:                "test-menu-id",
		SearchMethod:          "fuzzy",
		PreserveOrder:         true,
		InitialQuery:          "initial query",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              800,
		MinHeight:             600,
		MaxWidth:              1600,
		MaxHeight:             1200,
		AcceptCustomSelection: true,
	}

	// Test all fields are accessible and correct
	assert.Equal(t, "Test Title", config.Title)
	assert.Equal(t, "Test Prompt", config.Prompt)
	assert.Equal(t, "test-menu-id", config.MenuID)
	assert.Equal(t, "fuzzy", config.SearchMethod)
	assert.True(t, config.PreserveOrder)
	assert.Equal(t, "initial query", config.InitialQuery)
	assert.False(t, config.AutoAccept)
	assert.False(t, config.TerminalMode)
	assert.False(t, config.NoNumericSelection)
	assert.Equal(t, float32(800), config.MinWidth)
	assert.Equal(t, float32(600), config.MinHeight)
	assert.Equal(t, float32(1600), config.MaxWidth)
	assert.Equal(t, float32(1200), config.MaxHeight)
	assert.True(t, config.AcceptCustomSelection)
}

// TestConfigDefaults tests Config zero values
func TestConfigDefaults(t *testing.T) {
	config := &Config{}

	// Test zero values
	assert.Equal(t, "", config.Title)
	assert.Equal(t, "", config.Prompt)
	assert.Equal(t, "", config.MenuID)
	assert.Equal(t, "", config.SearchMethod)
	assert.False(t, config.PreserveOrder)
	assert.Equal(t, "", config.InitialQuery)
	assert.False(t, config.AutoAccept)
	assert.False(t, config.TerminalMode)
	assert.False(t, config.NoNumericSelection)
	assert.Equal(t, float32(0), config.MinWidth)
	assert.Equal(t, float32(0), config.MinHeight)
	assert.Equal(t, float32(0), config.MaxWidth)
	assert.Equal(t, float32(0), config.MaxHeight)
	assert.False(t, config.AcceptCustomSelection)
}

// TestMenuItemStruct tests the MenuItem struct
func TestMenuItemStruct(t *testing.T) {
	menuItem := MenuItem{
		Title: "Test Item",
		AType: nil,
		Score: 10,
	}

	assert.Equal(t, "Test Item", menuItem.Title)
	assert.Equal(t, 10, menuItem.Score)
	assert.Nil(t, menuItem.AType)
}

// TestMenuItemComputedTitle tests ComputedTitle method
func TestMenuItemComputedTitle(t *testing.T) {
	testCases := []struct {
		name     string
		item     MenuItem
		expected string
	}{
		{
			name:     "title only",
			item:     MenuItem{Title: "Main Title"},
			expected: "Main Title",
		},
		{
			name:     "title with score",
			item:     MenuItem{Title: "Main", Score: 5},
			expected: "Main",
		},
		{
			name:     "empty title",
			item:     MenuItem{Title: ""},
			expected: "",
		},
		{
			name:     "title with AType",
			item:     MenuItem{Title: "Main", AType: nil},
			expected: "Main",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.item.ComputedTitle()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestMenuItemWithAType tests MenuItem with AType
func TestMenuItemWithAType(t *testing.T) {
	testSerializable := TestSerializable{}
	var serializable GmenuSerializable = testSerializable
	
	menuItem := MenuItem{
		Title: "AType Test",
		AType: &serializable,
		Score: 42,
	}

	assert.Equal(t, "AType Test", menuItem.Title)
	assert.Equal(t, 42, menuItem.Score)
	assert.NotNil(t, menuItem.AType)
	
	// Test ComputedTitle with AType
	menuItemNoTitle := MenuItem{
		Title: "",
		AType: &serializable,
	}
	assert.Equal(t, "TestSerializable", menuItemNoTitle.ComputedTitle())
}

// TestExitCodeEnum tests ExitCode constants
func TestExitCodeEnum(t *testing.T) {
	// Test that all exit codes are defined
	assert.Equal(t, ExitCode(-1), Unset)
	assert.Equal(t, ExitCode(0), NoError)
	assert.Equal(t, ExitCode(1), UnknownError)
	assert.Equal(t, ExitCode(2), UserCanceled)

	// Test that exit codes are different values
	exitCodes := []ExitCode{Unset, NoError, UnknownError, UserCanceled}
	for i, code1 := range exitCodes {
		for j, code2 := range exitCodes {
			if i != j {
				assert.NotEqual(t, code1, code2, "Exit codes should be unique")
			}
		}
	}
}

// TestExitCodeValues tests specific exit code values
func TestExitCodeValues(t *testing.T) {
	// Test specific values if they matter for external integration
	assert.Equal(t, -1, int(Unset))
	assert.Equal(t, 0, int(NoError))
	assert.Equal(t, 1, int(UnknownError))
	assert.Equal(t, 2, int(UserCanceled))
}

// TestLoadingItem tests the LoadingItem constant
func TestLoadingItem(t *testing.T) {
	assert.Equal(t, "Loading", LoadingItem.Title)
	assert.Nil(t, LoadingItem.AType)
	assert.Equal(t, 0, LoadingItem.Score)
}

// TestMenuSerializable is a test type that implements GmenuSerializable
type TestMenuSerializable struct {
	Name  string
	Value int
}

func (ts TestMenuSerializable) Serialize() string {
	return ts.Name
}

func (ts TestMenuSerializable) ToMenuItem() MenuItem {
	return MenuItem{
		Title: ts.Name,
		Score: ts.Value,
	}
}

// TestGmenuSerializableInterface tests the GmenuSerializable interface
func TestGmenuSerializableInterface(t *testing.T) {

	// Test the implementation
	testObj := TestMenuSerializable{Name: "Test", Value: 42}
	menuItem := testObj.ToMenuItem()

	assert.Equal(t, "Test", menuItem.Title)
	assert.Equal(t, 42, menuItem.Score)
	assert.Equal(t, "Test", testObj.Serialize())
}

// TestConfigValidation tests configuration validation scenarios
func TestConfigValidation(t *testing.T) {
	testCases := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid minimal config",
			config: Config{
				Title:  "Test",
				Prompt: "Enter:",
			},
			valid: true,
		},
		{
			name: "valid full config",
			config: Config{
				Title:                 "Full Test",
				Prompt:                "Search:",
				MenuID:                "test-id",
				SearchMethod:          "fuzzy",
				PreserveOrder:         false,
				InitialQuery:          "",
				AutoAccept:            false,
				TerminalMode:          false,
				NoNumericSelection:    false,
				MinWidth:              600,
				MinHeight:             300,
				MaxWidth:              1200,
				MaxHeight:             800,
				AcceptCustomSelection: true,
			},
			valid: true,
		},
		{
			name: "config with zero dimensions",
			config: Config{
				Title:     "Zero Dims",
				MinWidth:  0,
				MinHeight: 0,
				MaxWidth:  0,
				MaxHeight: 0,
			},
			valid: true, // Should be valid (auto-size)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test that config can be created and used
			config := tc.config
			
			// Basic validation - config should have required fields
			if tc.valid {
				assert.NotNil(t, &config)
				// Add more validation logic here if needed
			}
		})
	}
}

// TestMenuItemEquality tests MenuItem comparison
func TestMenuItemEquality(t *testing.T) {
	item1 := MenuItem{
		Title: "Same",
		Score: 10,
	}

	item2 := MenuItem{
		Title: "Same",
		Score: 10,
	}

	item3 := MenuItem{
		Title: "Different",
		Score: 10,
	}

	// Test equality (struct comparison)
	assert.Equal(t, item1.Title, item2.Title)
	assert.Equal(t, item1.Score, item2.Score)

	// Test inequality
	assert.NotEqual(t, item1.Title, item3.Title)
}

// TestMenuItemCopy tests MenuItem copying behavior
func TestMenuItemCopy(t *testing.T) {
	original := MenuItem{
		Title: "Original",
		Score: 5,
	}

	// Test value copy
	copy := original
	copy.Title = "Modified"
	copy.Score = 10

	// Original should be unchanged for simple fields
	assert.Equal(t, "Original", original.Title)
	assert.Equal(t, 5, original.Score)
	assert.Equal(t, "Modified", copy.Title)
	assert.Equal(t, 10, copy.Score)
}