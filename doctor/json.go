// Package doctor — JSON emitter for `movie doctor --json`.
// Output schema is stable for scripting/CI consumption.
//
// Naming: this file follows spec/01-coding-guidelines/.../09-acronym-naming.md
// (project-specific MixedCaps rule). `JSON` becomes `Json` in identifiers;
// the user-facing flag and JSON wire format are unchanged.
package doctor

import (
	"encoding/json"
	"fmt"
)

// JsonReport is the stable wire-format for `movie doctor --json`.
// Field names are snake_case for cross-language consumers.
// Field order optimized for govet fieldalignment (strings, slice, bools last).
type JsonReport struct {
	Schema    string        `json:"schema"`
	Source    string        `json:"deploy_source"`
	Target    string        `json:"active_binary"`
	DeployDir string        `json:"deploy_dir"`
	Findings  []JsonFinding `json:"findings"`
	Repo      JsonRepo      `json:"repo"`
	HasErr    bool          `json:"has_errors"`
	HasFix    bool          `json:"has_fixable"`
}

// JsonFinding mirrors Finding with snake_case JSON tags.
type JsonFinding struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Severity  string `json:"severity"`
	Detail    string `json:"detail"`
	FixHint   string `json:"fix_hint"`
	IsFixable bool   `json:"is_fixable"`
}

// JsonRepo is the wire-format for the one-line repo staleness summary.
type JsonRepo struct {
	Branch    string `json:"branch"`
	Summary   string `json:"summary"`
	Recovery  string `json:"recovery"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
	IsGitRepo bool   `json:"is_git_repo"`
	IsClean   bool   `json:"is_clean"`
	IsCurrent bool   `json:"is_current"`
}

// PrintJson writes the report as indented JSON to stdout.
func (r *Report) PrintJson() error {
	payload := r.toJson()
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func (r *Report) toJson() JsonReport {
	return JsonReport{
		Schema:    "movie-doctor/v1",
		Source:    r.Source,
		Target:    r.Target,
		DeployDir: r.DeployDir,
		HasErr:    r.HasErrors(),
		HasFix:    r.HasFixable(),
		Findings:  toJsonFindings(r.Findings),
		Repo:      toJsonRepo(r.Repo),
	}
}

func toJsonRepo(s RepoStatus) JsonRepo {
	return JsonRepo{
		Branch:    s.Branch,
		Summary:   s.Summary,
		Recovery:  s.Recovery,
		Ahead:     s.Ahead,
		Behind:    s.Behind,
		IsGitRepo: s.IsGitRepo,
		IsClean:   s.IsClean,
		IsCurrent: s.IsCurrent,
	}
}

func toJsonFindings(findings []Finding) []JsonFinding {
	out := make([]JsonFinding, 0, len(findings))
	for _, f := range findings {
		out = append(out, JsonFinding{
			ID:        f.ID,
			Title:     f.Title,
			Severity:  string(f.Severity),
			Detail:    f.Detail,
			FixHint:   f.FixHint,
			IsFixable: f.IsFixable,
		})
	}
	return out
}
