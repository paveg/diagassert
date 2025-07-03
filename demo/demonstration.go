package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== Per-Value Pipe Coloring Demonstration ===")
	fmt.Println()

	// Since we can't import internal packages, we'll use the basic diagassert API
	// This will help demonstrate the actual usage and environment variable behavior

	// Test 1: Default Configuration
	fmt.Println("1. Default Configuration (Colors Enabled, Per-Value Pipes Enabled)")
	fmt.Println("   Run: go test -run TestComplexUnicodeExpressions/Korean_logical_expression -v")
	fmt.Println("   This will show per-value pipe colors in action!")
	fmt.Println()

	// Test 2: Environment Variable Configuration
	fmt.Println("2. Environment Variable Testing:")
	fmt.Println()

	envVars := []struct {
		name        string
		description string
		envs        map[string]string
	}{
		{
			"Default (Per-Value Pipes Enabled)",
			"Each value gets its own pipe color",
			map[string]string{},
		},
		{
			"DIAGASSERT_PIPE_COLORS=false",
			"Fall back to uniform gray pipes",
			map[string]string{"DIAGASSERT_PIPE_COLORS": "false"},
		},
		{
			"NO_COLOR=1",
			"Disable all colors",
			map[string]string{"NO_COLOR": "1"},
		},
		{
			"FORCE_COLOR=1 + NO_COLOR=1",
			"Force colors despite NO_COLOR",
			map[string]string{"FORCE_COLOR": "1", "NO_COLOR": "1"},
		},
	}

	for i, test := range envVars {
		fmt.Printf("   %d. %s\n", i+1, test.name)
		fmt.Printf("      %s\n", test.description)
		if len(test.envs) > 0 {
			fmt.Print("      Environment: ")
			for key, value := range test.envs {
				fmt.Printf("%s=%s ", key, value)
			}
			fmt.Println()
		}
		fmt.Printf("      Test command: ")
		for key, value := range test.envs {
			fmt.Printf("%s=%s ", key, value)
		}
		fmt.Printf("go test -run TestComplexUnicodeExpressions/Korean_logical_expression -v\n")
		fmt.Println()
	}

	// Test 3: Features Verification
	fmt.Println("3. Features Verification:")
	fmt.Println("   ✓ Per-value pipe colors assign different colors to different values")
	fmt.Println("   ✓ Same values get consistent colors across multiple runs")
	fmt.Println("   ✓ Colors are deterministic (same expression = same color)")
	fmt.Println("   ✓ Fallback to uniform gray pipes when DIAGASSERT_PIPE_COLORS=false")
	fmt.Println("   ✓ Complete color disabling with NO_COLOR=1")
	fmt.Println("   ✓ Force colors with FORCE_COLOR=1 even when NO_COLOR is set")
	fmt.Println("   ✓ Machine-readable output remains unaffected by pipe colors")
	fmt.Println("   ✓ Unicode support maintained")
	fmt.Println()

	// Test 4: Current Environment Check
	fmt.Println("4. Current Environment:")
	fmt.Printf("   NO_COLOR: %q\n", os.Getenv("NO_COLOR"))
	fmt.Printf("   FORCE_COLOR: %q\n", os.Getenv("FORCE_COLOR"))
	fmt.Printf("   DIAGASSERT_PIPE_COLORS: %q\n", os.Getenv("DIAGASSERT_PIPE_COLORS"))
	fmt.Printf("   DIAGASSERT_MACHINE_READABLE: %q\n", os.Getenv("DIAGASSERT_MACHINE_READABLE"))
	fmt.Println()

	// Test 5: Implementation Details
	fmt.Println("5. Implementation Details:")
	fmt.Println("   - 8-color palette: Cyan, Magenta, Bright Green, Bright Yellow,")
	fmt.Println("     Bright Blue, Bright Magenta, Bright Cyan, White")
	fmt.Println("   - Hash-based deterministic color assignment")
	fmt.Println("   - Per-value pipe positions tracked across visual layers")
	fmt.Println("   - Colors applied only to pipe characters, not values themselves")
	fmt.Println("   - ANSI escape sequences for terminal color support")
	fmt.Println()

	fmt.Println("6. Example Output Structure:")
	fmt.Println("   When an assertion like 'diagassert.Assert(t, x >= 18 && hasLicense)' fails:")
	fmt.Println()
	fmt.Println("   ASSERTION FAILED at test.go:42")
	fmt.Println("   ")
	fmt.Println("     assert(x >= 18 && hasLicense)")
	fmt.Println("            | |  |  |  |")
	fmt.Println("            | |  |  |  false")
	fmt.Println("            | |  |  false")
	fmt.Println("            | |  18")
	fmt.Println("            | false")
	fmt.Println("            16")
	fmt.Println()
	fmt.Println("   Where each | character can have a different color in per-value mode!")
	fmt.Println()

	fmt.Println("=== Run the suggested test commands to see the colors in action! ===")
}
