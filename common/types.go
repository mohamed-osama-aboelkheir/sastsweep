package common

type Options struct {
	// Target options
	Repo  string
	Repos string

	// Running options
	Fireprox   string
	Debug      bool
	Threads    int
	OutDir     string
	SaveRepo   bool
	NoEmoji   bool
	ConfigPath string

	// Display options
	Description     bool
	FullDescription bool
	Stars           bool
	Language        bool
	Topics          bool
	Files           bool
	RepoLink        bool
	Forks           bool
	LastRelease     bool
	LastCommit      bool
	FirstCommit     bool
	Commits         bool
	SecurityIssues  bool
	Watchers        bool
	Contributors    bool
	Branch          bool
	Issues          bool
	PullRequests    bool

	// Filters
	FilterStars          string
	FilterFiles          string
	FilterForks          string
	FilterLastRelease    string
	FilterLastCommit     string
	FilterFirstCommit    string
	FilterCommits        string
	FilterSecurityIssues string
	FilterWatchers       string
	FilterContributors   string
	FilterIssues         string
	FilterPullRequests   string

	// Scanning options
	NoSemgrep   bool
	SemgrepPath string

	// Reporting options
	Github1s bool
	RawLinks bool
}
