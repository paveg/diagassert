package testutil

import (
	"fmt"
	"strings"
)

// MockT is a mock implementation of the TestingT interface
type MockT struct {
	failed   bool
	messages []string
}

func NewMockT() *MockT {
	return &MockT{
		messages: make([]string, 0),
	}
}

func (m *MockT) Fatal(args ...interface{}) {
	for _, arg := range args {
		m.messages = append(m.messages, fmt.Sprint(arg))
	}
	m.failed = true
	panic("FailNow called")
}

func (m *MockT) Error(args ...interface{}) {
	for _, arg := range args {
		m.messages = append(m.messages, fmt.Sprint(arg))
	}
	m.failed = true
}

func (m *MockT) Helper() {}

func (m *MockT) Failed() bool {
	return m.failed
}

func (m *MockT) GetOutput() string {
	return strings.Join(m.messages, "\n")
}

// Test struct
type User struct {
	Name string
	Age  int
}

func (u User) IsAdult() bool {
	return u.Age >= 18
}

func (u User) HasLicense() bool {
	return false // Always returns false
}
