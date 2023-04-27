package scriptrunner

import (
	"testing"
	"time"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestSetLastRuntime(t *testing.T) {
	err := setLastRuntime("unit_test_monitor", time.Now())
	if err != nil {
		t.Fatalf("setLastRuntime failed: %v", err)
	}
}
