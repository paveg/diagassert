package formatter

import (
	"os"
	"testing"

	"github.com/fatih/color"
)

func TestColorDebug(t *testing.T) {
	// Test 1: Default behavior
	t.Run("default", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		os.Unsetenv("FORCE_COLOR")
		color.NoColor = false

		enabled := shouldEnableColors()
		t.Logf("Default: shouldEnableColors() = %v, color.NoColor = %v", enabled, color.NoColor)

		formatter := NewVisualFormatter()
		t.Logf("Formatter ColorsEnabled = %v", formatter.colorConfig.ColorsEnabled)
	})

	// Test 2: NO_COLOR only
	t.Run("NO_COLOR only", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		os.Unsetenv("FORCE_COLOR")
		color.NoColor = false

		enabled := shouldEnableColors()
		t.Logf("NO_COLOR only: shouldEnableColors() = %v, color.NoColor = %v", enabled, color.NoColor)

		formatter := NewVisualFormatter()
		t.Logf("Formatter ColorsEnabled = %v", formatter.colorConfig.ColorsEnabled)
	})

	// Test 3: FORCE_COLOR only
	t.Run("FORCE_COLOR only", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		os.Setenv("FORCE_COLOR", "1")
		color.NoColor = false

		enabled := shouldEnableColors()
		t.Logf("FORCE_COLOR only: shouldEnableColors() = %v, color.NoColor = %v", enabled, color.NoColor)

		formatter := NewVisualFormatter()
		t.Logf("Formatter ColorsEnabled = %v", formatter.colorConfig.ColorsEnabled)
	})

	// Test 4: Both NO_COLOR and FORCE_COLOR
	t.Run("both NO_COLOR and FORCE_COLOR", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		os.Setenv("FORCE_COLOR", "1")
		color.NoColor = false

		enabled := shouldEnableColors()
		t.Logf("Both: shouldEnableColors() = %v, color.NoColor = %v", enabled, color.NoColor)

		formatter := NewVisualFormatter()
		t.Logf("Formatter ColorsEnabled = %v", formatter.colorConfig.ColorsEnabled)

		// Test if colors actually work
		red := color.New(color.FgRed)
		output := red.Sprint("test")
		t.Logf("Red color output: %q", output)
	})

	// Clean up
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("FORCE_COLOR")
}
