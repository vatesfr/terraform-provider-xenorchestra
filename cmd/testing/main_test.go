package main

import "testing"

func Test_createGoTestArgs(t *testing.T) {
	tests := []string{"one", "two", "three"}
	args := createGoTestArgs(tests, false)

	if len(args) < 6 || args[5] != `-run='one\|two\|three'` {
		t.Errorf("failed to template command with specific tests actual: %v", args[4])
	}
}

func Test_findRemainingTests(t *testing.T) {
	successfulTests := []string{"one", "two", "four"}
	remainingTests := []string{"one", "three"}

	remaining := findRemainingTests(successfulTests, remainingTests)
	if len(remaining) != 1 || remaining[0] != "three" {
		t.Errorf("expected remaining to contain 'three'")
	}
}
