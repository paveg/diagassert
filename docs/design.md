# diagassert - Detailed Design and TDD Plan

## API Design

```go
package diagassert

// Basic assertion functions
func Assert(t testing.TB, expr bool, msgAndArgs ...interface{})
func Require(t testing.TB, expr bool, msgAndArgs ...interface{})

// Version with specified output format
func AssertWithFormat(t testing.TB, expr bool, format OutputFormat, msgAndArgs ...interface{})

// Assertion with options
func AssertWithOptions(t testing.TB, expr bool, opts ...Option)

// Options
type Option func(*assertOptions)

func WithContext(key string, value interface{}) Option
func WithVariableHistory(varName string) Option  
func WithSnapshot(varName string, value interface{}) Option
func WithFormat(format OutputFormat) Option

// Output formats
type OutputFormat int

const (
    FormatHybrid   OutputFormat = iota // Default
    FormatHuman                        // Human-only
    FormatMachine                      // Machine-only  
    FormatMarkdown                     // Markdown format
    FormatJSON                         // JSON format
)

// Global configuration
func SetDefaultFormat(format OutputFormat)
func EnableMachineReadable()   // Set to FormatHybrid
func DisableMachineReadable()  // Set to FormatHuman
```

## Hybrid Output Format Specification

### On Success

```
ASSERTION PASSED at user_test.go:42
Expression: user.Age >= 18
Result: true
```

### On Failure

```
ASSERTION FAILED at user_test.go:42
Expression: user.Age >= 18 && user.HasLicense()
Result: false

EVALUATION TRACE:
├─ user.Age >= 18                    [ID:1,TYPE:comparison]
│  ├─ user.Age = 16 (int)           [VAR:user.Age,CHANGED:false]
│  ├─ 18 = 18 (int)                 [CONST]
│  └─ RESULT: false                  [BRANCH:left]
└─ user.HasLicense()                 [ID:2,TYPE:method_call]
   ├─ user = User{Name:"Alice"...}   [VAR:user,TYPE:User]
   └─ RESULT: false                  [BRANCH:right]

CONTEXT:
- test_case: "underage user"
- previous_age: 15

[MACHINE_READABLE_START]
EXPR_TREE: ((1:comparison:false) AND (2:method_call:false))
VAR_STATES: user.Age=16,user.Name="Alice"
FAIL_REASON: left_operand_false
FAIL_PATH: 1->user.Age >= 18
[MACHINE_READABLE_END]
```