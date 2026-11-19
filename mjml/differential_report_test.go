package mjml

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type differentialReport struct {
	Reference  string          `json:"reference"`
	Sources    []sourceSummary `json:"sources"`
	Total      sourceSummary   `json:"total"`
	Clusters   []clusterReport `json:"clusters"`
	Validation []tally         `json:"validation"`
}

type sourceSummary struct {
	Name          string `json:"name"`
	Licence       string `json:"licence,omitempty"`
	Skipped       string `json:"skipped,omitempty"`
	Cases         int    `json:"cases"`
	Equivalent    int    `json:"equivalent"`
	Differing     int    `json:"differing"`
	ByteIdentical int    `json:"byteIdentical"`
	BothReject    int    `json:"bothReject"`
	Crashes       int    `json:"crashes"`
}

type clusterReport struct {
	ID          string         `json:"id"`
	Reason      string         `json:"reason"`
	Cases       int            `json:"cases"`
	BySource    map[string]int `json:"bySource"`
	Signatures  []string       `json:"signatures"`
	Examples    []string       `json:"examples"`
	Example     string         `json:"example"`
	Differences string         `json:"differences"`
	Diff        string         `json:"diff"`
}

// tally counts a gomjml validation error that MJML does not report.
type tally struct {
	Message  string   `json:"message"`
	Cases    int      `json:"cases"`
	Examples []string `json:"examples"`
}

func (run *differentialRun) report() differentialReport {
	report := differentialReport{Reference: "mjml 4.16.1 (mjml-4.17 2.x for mj-wrapper-gap)"}
	report.Total.Name = "total"
	members := map[string][]*caseResult{}
	validation := map[string][]string{}
	for _, source := range run.sources {
		summary := sourceSummary{Name: source.source.name, Licence: source.source.licence, Skipped: source.skipped}
		for _, result := range source.results {
			summary.Cases++
			switch {
			case result.failing():
				summary.Differing++
			default:
				summary.Equivalent++
			}
			if result.byteIdentical {
				summary.ByteIdentical++
			}
			if result.referenceError != "" && result.gomjml == "error" {
				summary.BothReject++
			}
			if result.crash != "" {
				summary.Crashes++
			}
			clusters, _ := run.clustersOf(result)
			for _, id := range clusters {
				members[id] = append(members[id], result)
			}
			for _, v := range result.validation {
				validation[v] = append(validation[v], result.name)
			}
		}
		report.Sources = append(report.Sources, summary)
		report.Total.Cases += summary.Cases
		report.Total.Equivalent += summary.Equivalent
		report.Total.Differing += summary.Differing
		report.Total.ByteIdentical += summary.ByteIdentical
		report.Total.BothReject += summary.BothReject
		report.Total.Crashes += summary.Crashes
	}
	slices.SortFunc(report.Sources, func(a, b sourceSummary) int {
		return slices.IndexFunc(differentialSources, func(s diffSource) bool { return s.name == a.Name }) -
			slices.IndexFunc(differentialSources, func(s diffSource) bool { return s.name == b.Name })
	})

	for _, cluster := range run.clusters {
		results := members[cluster.ID]
		if len(results) == 0 {
			continue
		}
		slices.SortFunc(results, func(a, b *caseResult) int { return strings.Compare(a.name, b.name) })
		c := clusterReport{ID: cluster.ID, Reason: cluster.Reason, Cases: len(results), BySource: map[string]int{}, Signatures: cluster.Signatures}
		for _, result := range results {
			source, _, _ := strings.Cut(result.name, "/")
			c.BySource[source]++
		}
		// The clearest example shows the fewest other differences.
		example := slices.MinFunc(results, func(a, b *caseResult) int {
			if n := len(a.signatures) - len(b.signatures); n != 0 {
				return n
			}
			return len(a.name) - len(b.name)
		})
		c.Example = example.name
		c.Examples = append(c.Examples, example.name)
		for _, result := range results {
			if len(c.Examples) >= 6 {
				break
			}
			if result != example {
				c.Examples = append(c.Examples, result.name)
			}
		}
		c.Differences = trimLines(ansiEscape.ReplaceAllString(strings.Join(example.differences, "\n\n"), ""), 30)
		c.Diff = caseDiff(example, 2, 40, signatureWords(cluster.Signatures)...)
		report.Clusters = append(report.Clusters, c)
	}
	slices.SortStableFunc(report.Clusters, func(a, b clusterReport) int {
		if a.Cases != b.Cases {
			return b.Cases - a.Cases
		}
		return strings.Compare(a.ID, b.ID)
	})

	for _, message := range slices.Sorted(maps.Keys(validation)) {
		cases := validation[message]
		report.Validation = append(report.Validation, tally{message, len(cases), cases[:min(len(cases), 3)]})
	}
	slices.SortStableFunc(report.Validation, func(a, b tally) int { return b.Cases - a.Cases })
	return report
}

var signatureWord = regexp.MustCompile(`[{\[](.+?)[}\]]|element ([\w:-]+)|tag </?([\w:-]+)|\.([a-z][\w-]*)`)

// signatureWords are the property, attribute, tag and class names a cluster's signatures
// mention, for picking the diff hunks that show it.
func signatureWords(signatures []string) []string {
	var words []string
	for _, signature := range signatures {
		for _, match := range signatureWord.FindAllStringSubmatch(signature, -1) {
			for _, group := range match[1:] {
				for word := range strings.SplitSeq(group, ",") {
					if word = strings.TrimSuffix(word, "-N"); word != "" {
						words = append(words, word)
					}
				}
			}
		}
	}
	slices.Sort(words)
	return slices.Compact(words)
}

func trimLines(text string, n int) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= n {
		return text
	}
	return strings.Join(lines[:n], "\n") + fmt.Sprintf("\n… (%d more lines)", len(lines)-n)
}

func writeReport(dir string, report differentialReport) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.json"), append(content, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "report.md"), []byte(report.markdown()), 0o644)
}

func percent(n, of int) string {
	if of == 0 {
		return "–"
	}
	return fmt.Sprintf("%.1f%%", 100*float64(n)/float64(of))
}

func (report differentialReport) markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# gomjml differential report\n\nReference: %s, rendered as `mjml --config.beautify false --config.minify true` would.\n\n", report.Reference)
	b.WriteString("The verdict is semantic: the comparison of `TestMJMLAgainstExpected` (DOM, text order, inline-style declarations, " +
		"style blocks in order, MSO conditionals and VML, URLs), which ignores attribute order, collapsible whitespace and generated id values. " +
		"Byte-identical counts outputs that are equal byte for byte once ids are masked; it is reported, not gated.\n\n")
	b.WriteString("| source | licence | cases | equivalent | differing | byte-identical | both reject | crashes |\n|---|---|---:|---:|---:|---:|---:|---:|\n")
	for _, s := range append(slices.Clone(report.Sources), report.Total) {
		cases := fmt.Sprint(s.Cases)
		if s.Skipped != "" {
			cases = "skipped: " + s.Skipped
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %d (%s) | %d | %d (%s) | %d | %d |\n", s.Name, s.Licence, cases,
			s.Equivalent, percent(s.Equivalent, s.Cases), s.Differing, s.ByteIdentical, percent(s.ByteIdentical, s.Cases), s.BothReject, s.Crashes)
	}

	fmt.Fprintf(&b, "\n## Defect clusters (%d)\n\nA cluster is a set of difference signatures (`component kind target`) with one cause. A case can show several.\n\n", len(report.Clusters))
	b.WriteString("| # | cluster | cases | sources | reason |\n|---:|---|---:|---|---|\n")
	for i, c := range report.Clusters {
		var sources []string
		for _, source := range slices.Sorted(maps.Keys(c.BySource)) {
			sources = append(sources, fmt.Sprintf("%s %d", source, c.BySource[source]))
		}
		fmt.Fprintf(&b, "| %d | [%s](#%s) | %d | %s | %s |\n", i+1, c.ID, c.ID, c.Cases, strings.Join(sources, ", "), c.Reason)
	}
	for _, c := range report.Clusters {
		fmt.Fprintf(&b, "\n### %s\n\n%s\n\n**%d cases.** Examples: %s.\n\nSignatures:\n\n", c.ID, orDash(c.Reason), c.Cases, "`"+strings.Join(c.Examples, "`, `")+"`")
		for i, signature := range c.Signatures {
			if i == 12 {
				fmt.Fprintf(&b, "- … and %d more\n", len(c.Signatures)-12)
				break
			}
			fmt.Fprintf(&b, "- `%s`\n", signature)
		}
		fmt.Fprintf(&b, "\nSemantic differences in `%s`:\n\n```\n%s\n```\n\nDiff of `%s` (`go test ./mjml -run '^TestDifferential$' -args -diff.case=%s` prints it in full):\n\n```diff\n%s```\n",
			c.Example, c.Differences, c.Example, c.Example, c.Diff)
	}

	if len(report.Validation) > 0 {
		b.WriteString("\n## Validation errors only gomjml reports\n\ngomjml's `Render` returns these as an error alongside the HTML; MJML's soft validation reports nothing for the tag on that line. Not gated.\n\n| error | cases | examples |\n|---|---:|---|\n")
		for _, v := range report.Validation {
			fmt.Fprintf(&b, "| %s | %d | %s |\n", strings.ReplaceAll(v.Message, "|", `\|`), v.Cases, "`"+strings.Join(v.Examples, "`, `")+"`")
		}
	}
	return b.String()
}

func orDash(s string) string {
	if s == "" {
		return "_No reason yet: triage this cluster._"
	}
	return s
}
