// report.go — pretty-printer for the doctor report.
// Output uses the v2.123.0 indent scheme: 0/1/2/3 spaces with bracketed tags.
package doctor

import (
	"fmt"
	"strings"
)

const (
	tagOK   = "[ OK ]"
	tagWarn = "[WARN]"
	tagErr  = "[ERR ]"
)

// Print writes the human-readable report to stdout.
func (r *Report) Print() {
	fmt.Println("==> movie doctor")
	fmt.Println("  --------------------------------------------------")
	printPathSummary(r)
	fmt.Println("  --------------------------------------------------")
	for _, f := range r.Findings {
		printFinding(f)
	}
	printFooter(r)
}

func printPathSummary(r *Report) {
	fmt.Printf("    deploy source : %s\n", orDash(r.Source))
	fmt.Printf("    active binary : %s\n", orDash(r.Target))
	fmt.Printf("    deploy dir    : %s\n", orDash(r.DeployDir))
}

func printFinding(f Finding) {
	tag := tagFor(f.Severity)
	fmt.Printf("    %s %s\n", tag, f.Title)
	if f.Detail != "" {
		for _, line := range strings.Split(f.Detail, "\n") {
			fmt.Printf("      %s\n", line)
		}
	}
	if f.FixHint != "" && f.Severity != SeverityOK {
		fmt.Printf("      hint: %s\n", f.FixHint)
	}
}

func printFooter(r *Report) {
	fmt.Println("  --------------------------------------------------")
	if r.HasErrors() {
		fmt.Println("  Result: errors found. Run `movie doctor --fix` to attempt repair.")
		return
	}
	if r.HasFixable() {
		fmt.Println("  Result: warnings found. Run `movie doctor --fix` to clean up.")
		return
	}
	fmt.Println("  Result: all good.")
}

func tagFor(sev Severity) string {
	switch sev {
	case SeverityOK:
		return tagOK
	case SeverityErr:
		return tagErr
	default:
		return tagWarn
	}
}

func orDash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}
