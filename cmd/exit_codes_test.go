// exit_codes_test.go — locks the public CI contract for undo/redo
// exit codes. The numeric values are part of the user-facing API.
package cmd

import (
	"strings"
	"testing"
)

func TestExitCodeValues(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"ExitOK", ExitOK, 0},
		{"ExitGenericError", ExitGenericError, 2},
		{"ExitScopeRejected", ExitScopeRejected, 10},
		{"ExitRowDeclined", ExitRowDeclined, 11},
		{"ExitNothingMatched", ExitNothingMatched, 20},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d (CI contract!)", c.name, c.got, c.want)
		}
	}
}

func TestExitLabels(t *testing.T) {
	cases := map[int]string{
		ExitOK:             "ok",
		ExitGenericError:   "error",
		ExitScopeRejected:  "scope rejected",
		ExitRowDeclined:    "row declined",
		ExitNothingMatched: "nothing matched",
		999:                "unknown",
	}
	for code, want := range cases {
		if got := exitLabel(code); got != want {
			t.Errorf("exitLabel(%d) = %q, want %q", code, got, want)
		}
	}
}

// TestExitWithCodeSilentOnSuccess proves a 0-code call neither prints
// nor exits — required so happy-path runs don't add noise.
func TestExitWithCodeSilentOnSuccess(t *testing.T) {
	captured := ""
	originalPrint := exitFootPrintf
	originalExit := osExit
	exitFootPrintf = func(format string, a ...interface{}) { captured = format }
	osExit = func(int) { t.Fatal("osExit called for ExitOK — must be silent") }
	defer func() { exitFootPrintf = originalPrint; osExit = originalExit }()

	exitWithCode(ExitOK)
	if captured != "" {
		t.Errorf("ExitOK printed footer %q; expected silence", captured)
	}
}

func TestExitWithCodePropagatesNonZero(t *testing.T) {
	captured := ""
	exited := -1
	originalPrint := exitFootPrintf
	originalExit := osExit
	exitFootPrintf = func(format string, a ...interface{}) {
		captured = format
		_ = a
	}
	osExit = func(c int) { exited = c }
	defer func() { exitFootPrintf = originalPrint; osExit = originalExit }()

	exitWithCode(ExitScopeRejected)
	if exited != ExitScopeRejected {
		t.Errorf("osExit got %d, want %d", exited, ExitScopeRejected)
	}
	if !strings.Contains(captured, "exit:") {
		t.Errorf("footer %q missing 'exit:' prefix", captured)
	}
}
