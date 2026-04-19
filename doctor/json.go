// json.go — JSON emitter for `movie doctor --json`.
// Output schema is stable for scripting/CI consumption.
package doctor

import (
	"encoding/json"
	"fmt"
)

// JSONReport is the stable wire-format for `movie doctor --json`.
// Field names are snake_case for cross-language consumers.
type JSONReport struct {
	Schema    string         `json:"schema"`
	Source    string         `json:"deploy_source"`
	Target    string         `json:"active_binary"`
	DeployDir string         `json:"deploy_dir"`
	HasErr    bool           `json:"has_errors"`
	HasFix    bool           `json:"has_fixable"`
	Findings  []JSONFinding  `json:"findings"`
}

// JSONFinding mirrors Finding with snake_case JSON tags.
type JSONFinding struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Severity  string `json:"severity"`
	Detail    string `json:"detail"`
	FixHint   string `json:"fix_hint"`
	IsFixable bool   `json:"is_fixable"`
}

// PrintJSON writes the report as indented JSON to stdout.
func (r *Report) PrintJSON() error {
	payload := r.toJSON()
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func (r *Report) toJSON() JSONReport {
	return JSONReport{
		Schema:    "movie-doctor/v1",
		Source:    r.Source,
		Target:    r.Target,
		DeployDir: r.DeployDir,
		HasErr:    r.HasErrors(),
		HasFix:    r.HasFixable(),
		Findings:  toJSONFindings(r.Findings),
	}
}

func toJSONFindings(findings []Finding) []JSONFinding {
	out := make([]JSONFinding, 0, len(findings))
	for _, f := range findings {
		out = append(out, JSONFinding{
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
