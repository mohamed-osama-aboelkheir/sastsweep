package report

type SemgrepFinding struct {
	VulnerabilityTitle string
	Severity           string
	Description        string
	Code               string
	StartLine          int
	StopLine           int
	GithubLink         string
}

type ReportData struct {
	Target                     string
	VulnerabilityStats         map[string]int
	VulnerabilityStatsOrdering []string
	SeverityStats              map[string]int
	SeverityStatsOrdering      []string
	Findings                   []SemgrepFinding
}
