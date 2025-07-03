package diagassert_test

import (
	"testing"

	"github.com/paveg/diagassert"
)

// TestPowerAssertDemo_SimpleExpressions demonstrates basic power-assert output
// with proper alignment for simple expressions
func TestPowerAssertDemo_SimpleExpressions(t *testing.T) {
	t.Skip("Demo test - enable to see power-assert output")

	// Expected output:
	// assert_test.go:20
	//     diagassert.Assert(t, x > y)
	//                          | | |
	//                          | | 20
	//                          | false
	//                          10
	x := 10
	y := 20
	diagassert.Assert(t, x > y)

	// Expected output:
	// assert_test.go:30
	//     diagassert.Assert(t, a == b)
	//                          |  | |
	//                          |  | "world"
	//                          |  false
	//                          "hello"
	a := "hello"
	b := "world"
	diagassert.Assert(t, a == b)

	// Expected output:
	// assert_test.go:40
	//     diagassert.Assert(t, len(slice) > 5)
	//                          |   |      | |
	//                          |   |      | 5
	//                          |   |      false
	//                          3   [1 2 3]
	slice := []int{1, 2, 3}
	diagassert.Assert(t, len(slice) > 5)
}

// TestPowerAssertDemo_ComplexExpressions demonstrates nested expressions
// with depth-based display and proper tree structure
func TestPowerAssertDemo_ComplexExpressions(t *testing.T) {
	t.Skip("Demo test - enable to see power-assert output")

	// Expected output:
	// assert_test.go:55
	//     diagassert.Assert(t, (x > 0 && y < 10) || z == 100)
	//                          ||   |  | |   |   | |  |  |
	//                          ||   |  | |   |   | |  |  100
	//                          ||   |  | |   |   | |  false
	//                          ||   |  | |   |   | 50
	//                          ||   |  | |   |   false
	//                          ||   |  | |   20
	//                          ||   |  | false
	//                          ||   |  5
	//                          ||   true
	//                          |false
	//                          false
	x := 5
	y := 20
	z := 50
	diagassert.Assert(t, (x > 0 && y < 10) || z == 100)

	// Expected output:
	// assert_test.go:70
	//     diagassert.Assert(t, user.Age >= 18 && user.Name != "")
	//                          |    |   |  |   | |    |    |  |
	//                          |    |   |  |   | |    |    |  ""
	//                          |    |   |  |   | |    |    false
	//                          |    |   |  |   | |    "John"
	//                          |    |   |  |   | {Name:John Age:16}
	//                          |    |   |  |   false
	//                          |    |   |  18
	//                          |    |   true
	//                          |    16
	//                          {Name:John Age:16}
	type User struct {
		Name string
		Age  int
	}
	user := User{Name: "John", Age: 16}
	diagassert.Assert(t, user.Age >= 18 && user.Name != "")
}

// TestPowerAssertDemo_UnicodeSupport demonstrates proper alignment
// with Unicode characters (CJK languages)
func TestPowerAssertDemo_UnicodeSupport(t *testing.T) {
	t.Skip("Demo test - enable to see power-assert output")

	// Expected output:
	// assert_test.go:95
	//     diagassert.Assert(t, 名前 == "太郎")
	//                          |   |  |
	//                          |   |  "太郎"
	//                          |   false
	//                          "花子"
	名前 := "花子"
	diagassert.Assert(t, 名前 == "太郎")

	// Expected output:
	// assert_test.go:105
	//     diagassert.Assert(t, 한국어 != "안녕하세요")
	//                          |     |  |
	//                          |     |  "안녕하세요"
	//                          |     false
	//                          "안녕하세요"
	한국어 := "안녕하세요"
	diagassert.Assert(t, 한국어 != "안녕하세요")

	// Expected output:
	// assert_test.go:115
	//     diagassert.Assert(t, len(中文) > 3)
	//                          |   |    | |
	//                          |   |    | 3
	//                          |   |    false
	//                          2   "你好"
	中文 := "你好"
	diagassert.Assert(t, len(中文) > 3)
}

// TestPowerAssertDemo_VariousTypes demonstrates output for different Go types
func TestPowerAssertDemo_VariousTypes(t *testing.T) {
	t.Skip("Demo test - enable to see power-assert output")

	// Struct comparison
	// Expected output:
	// assert_test.go:130
	//     diagassert.Assert(t, p1 == p2)
	//                          |  |  |
	//                          |  |  {X:20 Y:30}
	//                          |  false
	//                          {X:10 Y:20}
	type Point struct{ X, Y int }
	p1 := Point{10, 20}
	p2 := Point{20, 30}
	diagassert.Assert(t, p1 == p2)

	// Map operations
	// Expected output:
	// assert_test.go:140
	//     diagassert.Assert(t, scores["Alice"] > scores["Bob"])
	//                          |      |       | |      |
	//                          |      |       | |      "Bob"
	//                          |      |       | 95
	//                          |      |       false
	//                          |      "Alice"
	//                          90
	scores := map[string]int{"Alice": 90, "Bob": 95}
	diagassert.Assert(t, scores["Alice"] > scores["Bob"])

	// Slice operations
	// Expected output:
	// assert_test.go:150
	//     diagassert.Assert(t, nums[0] + nums[1] == nums[2])
	//                          |   |   | |   |   |  |   |
	//                          |   |   | |   |   |  |   3
	//                          |   |   | |   |   |  [1 2 3 4 5]
	//                          |   |   | |   |   false
	//                          |   |   | |   2
	//                          |   |   | [1 2 3 4 5]
	//                          |   |   3
	//                          |   1
	//                          [1 2 3 4 5]
	nums := []int{1, 2, 3, 4, 5}
	diagassert.Assert(t, nums[0]+nums[1] == nums[2])

	// Pointer comparison
	// Expected output:
	// assert_test.go:165
	//     diagassert.Assert(t, *ptr1 == *ptr2)
	//                          ||    |  ||
	//                          ||    |  |20
	//                          ||    |  0xc000012345
	//                          ||    false
	//                          |10
	//                          0xc000012340
	var val1, val2 = 10, 20
	ptr1, ptr2 := &val1, &val2
	diagassert.Assert(t, *ptr1 == *ptr2)
}

// TestPowerAssertDemo_VisualTreeStructure demonstrates the visual tree
// structure for deeply nested expressions
func TestPowerAssertDemo_VisualTreeStructure(t *testing.T) {
	t.Skip("Demo test - enable to see power-assert output")

	// Expected output with tree structure:
	// assert_test.go:180
	//     diagassert.Assert(t, (a && (b || c)) || (d && e))
	//                          ||  |  |  | |   |  |  | |
	//                          ||  |  |  | |   |  |  | false
	//                          ||  |  |  | |   |  |  5
	//                          ||  |  |  | |   |  false
	//                          ||  |  |  | |   |  4
	//                          ||  |  |  | |   false
	//                          ||  |  |  | false
	//                          ||  |  |  3
	//                          ||  |  true
	//                          ||  |  false
	//                          ||  |  2
	//                          ||  true
	//                          ||  1
	//                          |false
	//                          false
	a, b, c, d, e := 1, 2, 3, 4, 5
	diagassert.Assert(t, (a != 0 && (b == 0 || c == 0)) || (d == 0 && e == 0))

	// Expected output for method chaining:
	// assert_test.go:200
	//     diagassert.Assert(t, user.GetAge() > 18 && user.IsActive())
	//                          |    |        | |   | |    |
	//                          |    |        | |   | |    false
	//                          |    |        | |   | {Name:Test Age:25 Active:false}
	//                          |    |        | |   false
	//                          |    |        | 18
	//                          |    |        true
	//                          |    25
	//                          {Name:Test Age:25 Active:false}
	type DemoUser struct {
		Name   string
		Age    int
		Active bool
	}
	user := &DemoUser{Name: "Test", Age: 25, Active: false}
	diagassert.Assert(t, user.Age > 18 && user.Active)
}

// TestPowerAssertDemo_ComplexRealWorld demonstrates real-world scenarios
func TestPowerAssertDemo_ComplexRealWorld(t *testing.T) {
	t.Skip("Demo test - enable to see power-assert output")

	// Expected output for validation logic:
	// assert_test.go:235
	//     diagassert.Assert(t, email != "" && len(email) > 5 && contains(email, "@"))
	//                          |     |  |   | |   |      | |  | |        |      |
	//                          |     |  |   | |   |      | |  | |        |      "@"
	//                          |     |  |   | |   |      | |  | |        "test"
	//                          |     |  |   | |   |      | |  | false
	//                          |     |  |   | |   |      | |  "test"
	//                          |     |  |   | |   |      | 5
	//                          |     |  |   | |   |      false
	//                          |     |  |   | 4
	//                          |     |  |   | "test"
	//                          |     |  |   false
	//                          |     |  ""
	//                          |     true
	//                          "test"
	email := "test"
	diagassert.Assert(t, email != "" && len(email) > 5 && contains(email, "@"))

	// Expected output for range validation:
	// assert_test.go:255
	//     diagassert.Assert(t, value >= min && value <= max)
	//                          |     |  |   |  |     |  |
	//                          |     |  |   |  |     |  100
	//                          |     |  |   |  |     false
	//                          |     |  |   |  150
	//                          |     |  |   true
	//                          |     |  0
	//                          |     true
	//                          150
	value, min, max := 150, 0, 100
	diagassert.Assert(t, value >= min && value <= max)
}

// Helper function for email validation demo
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestPowerAssertDemo_RequireTermination demonstrates Require's immediate termination
func TestPowerAssertDemo_RequireTermination(t *testing.T) {
	t.Skip("Demo test - enable to see power-assert output")

	// Expected output:
	// assert_test.go:280
	//     diagassert.Require(t, config != nil)
	//                           |      |  |
	//                           |      |  <nil>
	//                           |      false
	//                           <nil>
	// [Test terminates here - subsequent assertions are not executed]
	var config *Config
	diagassert.Require(t, config != nil)

	// This line should never be reached
	t.Log("This should not be printed")
}

type Config struct {
	Host string
	Port int
}
