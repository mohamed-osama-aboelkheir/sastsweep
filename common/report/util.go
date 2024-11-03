package report

import (
	"sort"
	"strings"
)

func getLanguage(title string) string {
	return strings.Split(title, ".")[0]
}

func toLowerCase(s string) string {
	return strings.ToLower(s)
}

func SortFindings(findings []SemgrepFinding) {
	sort.SliceStable(findings, func(i, j int) bool {
		severityOrder := map[string]int{"HIGH": 1, "MEDIUM": 2, "LOW": 3}
		if severityOrder[findings[i].Severity] != severityOrder[findings[j].Severity] {
			return severityOrder[findings[i].Severity] < severityOrder[findings[j].Severity]
		}
		return findings[i].VulnerabilityTitle < findings[j].VulnerabilityTitle
	})
}

func CalculateSemgrepMetrics(findings []SemgrepFinding) (map[string]int, []string, map[string]int) {
	vulnerabilityStats := map[string]int{}
	vulnerabilityStatsOrdering := []string{}
	severityStats := map[string]int{}

	for _, finding := range findings {
		if _, ok := vulnerabilityStats[finding.VulnerabilityTitle]; !ok {
			vulnerabilityStats[finding.VulnerabilityTitle] = 0
			vulnerabilityStatsOrdering = append(vulnerabilityStatsOrdering, finding.VulnerabilityTitle)
		}
		vulnerabilityStats[finding.VulnerabilityTitle] += 1

		if _, ok := severityStats[finding.Severity]; !ok {
			severityStats[finding.Severity] = 0
		}

		severityStats[finding.Severity] += 1
	}

	return vulnerabilityStats, vulnerabilityStatsOrdering, severityStats
}
