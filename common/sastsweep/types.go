package sastsweep

import "time"

type CommitInfo struct {
	Date string `json:"date"`
}

type RepoInfo struct {
	// Display options
	Target          string
	Description     string
	FullDescription string
	Stars           int
	Language        string
	Topics          string
	Files           int
	RepoLink        string
	Forks           int
	LastRelease     time.Time
	LastCommit      time.Time
	FirstCommit     time.Time
	Commits         int
	SecurityIssues  int
	Watchers        int
	Contributors    int
	Branch          string
	Issues          int
	PullRequests    int
	SemgrepHits     int
	ReportPath      string
}

type SemgrepJson struct {
	Results []Result `json:"results"`
}

type Result struct {
	CheckID string    `json:"check_id"`
	End     Position  `json:"end"`
	Extra   ExtraInfo `json:"extra"`
	Path    string    `json:"path"`
	Start   Position  `json:"start"`
}

type Position struct {
	Line int `json:"line"`
}

type ExtraInfo struct {
	Lines    string   `json:"lines"`
	Message  string   `json:"message"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Impact string `json:"impact"`
}
