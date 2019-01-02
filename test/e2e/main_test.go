package e2e_test

import (
	"os"
	"testing"
)

type devnullLogger struct{}

func (_ devnullLogger) Logf(string, ...interface{}) {}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
