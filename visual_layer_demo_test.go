package diagassert_test

import (
	"testing"

	"github.com/paveg/diagassert"
)

// This file demonstrates the new visual layer architecture in action

func TestVisualLayerDemo_FixedPipePositions(t *testing.T) {
	// Skip demo test to avoid test failures - enable to see the new layer architecture
	t.Skip("Demo test - enable to see new visual layer architecture with fixed pipe positions")

	x := 10
	y := 20

	// With the new layer architecture, pipes stay fixed and values are distributed
	// across visual layers to avoid conflicts. This should show:
	//
	//   assert(x > y)
	//          | | |
	//          10  20
	//            |
	//            false
	//
	// Notice how the pipes remain at fixed positions even when values might conflict
	diagassert.Assert(t, x > y,
		diagassert.V("x", x),
		diagassert.V("y", y))
}

func TestVisualLayerDemo_ConflictResolution(t *testing.T) {
	// Skip demo test to avoid test failures - enable to see conflict resolution
	t.Skip("Demo test - enable to see how layer architecture resolves value conflicts")

	verylongvariablename := 100
	a := 5

	// The old system might try to center values under expressions, causing overlaps.
	// The new layer architecture assigns conflicting values to different layers:
	//
	//   assert(verylongvariablename > a)
	//          |                     |
	//          100                   5
	//                                |
	//                                false
	//
	// Values that would overlap are placed on separate layers
	diagassert.Assert(t, verylongvariablename > a,
		diagassert.V("verylongvariablename", verylongvariablename),
		diagassert.V("a", a))
}

func TestVisualLayerDemo_ComplexLayering(t *testing.T) {
	// Skip demo test to avoid test failures - enable to see complex layering
	t.Skip("Demo test - enable to see complex multi-layer distribution")

	age := 16
	hasDriversLicense := false
	hasPassport := true
	country := "US"

	// Complex expressions with multiple potential conflicts are handled
	// by distributing values across multiple visual layers while keeping
	// pipes at fixed positions:
	//
	//   assert(age >= 18 && hasDriversLicense && (hasPassport || country == "US"))
	//          |   |  |  |  |                  |  |          |  |       |
	//          16     18    false                 true          "US"    "US"
	//              |     |                        |            |      |
	//              false false                    true         |      true
	//                    |                                     |
	//                    false                                 true
	//                                                         |
	//                                                         false
	diagassert.Assert(t,
		age >= 18 && hasDriversLicense && (hasPassport || country == "US"),
		diagassert.V("age", age),
		diagassert.V("hasDriversLicense", hasDriversLicense),
		diagassert.V("hasPassport", hasPassport),
		diagassert.V("country", country))
}

func TestVisualLayerDemo_ValueAlignment(t *testing.T) {
	// Skip demo test to avoid test failures - enable to see proper value alignment
	t.Skip("Demo test - enable to see values aligned with their pipes")

	score := 85
	passing := 60

	// Values are aligned directly under their corresponding expressions,
	// not centered. This maintains the connection between pipes and values:
	//
	//   assert(score >= passing)
	//          |     |  |
	//          85       60
	//                |
	//                true
	//
	// Notice how "85" starts at the same position as "score", not centered under it
	diagassert.Assert(t, score >= passing,
		diagassert.V("score", score),
		diagassert.V("passing", passing))
}

func TestVisualLayerDemo_PipeContinuity(t *testing.T) {
	// Skip demo test to avoid test failures - enable to see pipe continuity
	t.Skip("Demo test - enable to see how pipes maintain continuity across layers")

	a, b, c := 10, 20, 30
	result := (a + b) * c

	// Pipes continue across all layers where values appear, maintaining
	// visual connection between expression elements and their values:
	//
	//   assert(result > 500)
	//          |      | |
	//          600      500
	//                 |
	//                 true
	//
	// Even with complex expressions, the visual hierarchy is preserved
	diagassert.Assert(t, result > 500,
		diagassert.V("result", result),
		diagassert.V("a", a),
		diagassert.V("b", b),
		diagassert.V("c", c))
}

func TestVisualLayerDemo_NoSmartCentering(t *testing.T) {
	// Skip demo test to avoid test failures - enable to see elimination of smart centering
	t.Skip("Demo test - enable to see how smart centering is eliminated for pipe alignment")

	veryLongVariableName := 42
	x := 1

	// The old system might try to center short values under long variable names,
	// which breaks pipe alignment. The new system aligns values with pipes:
	//
	//   assert(veryLongVariableName > x)
	//          |                    | |
	//          42                     1
	//                                 |
	//                                 false
	//
	// "42" starts at the same position as "veryLongVariableName", maintaining alignment
	diagassert.Assert(t, veryLongVariableName > x,
		diagassert.V("veryLongVariableName", veryLongVariableName),
		diagassert.V("x", x))
}