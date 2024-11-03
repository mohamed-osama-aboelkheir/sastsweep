package runner

import (
	"flag"
	"os"
	"path/filepath"
	"sastsweep/common"
	"sastsweep/common/logger"
	"strings"
)

func ParseOptions() *common.Options {

	options := &common.Options{}

	// Targets
	flag.StringVar(&options.Repo, "repo", "", "GitHub repository to scan")
	flag.StringVar(&options.Repos, "repos", "", "File of GitHub repositories to scan")

	// Running options
	flag.StringVar(&options.Fireprox, "fireprox", "", "Use fireprox for reasons... relates to rate limiting on a certain platform (ex: https://abcdefghi.execute-api.us-east-1.amazonaws.com/fireprox/)")
	flag.BoolVar(&options.Debug, "debug", false, "Enable debug messages")
	flag.IntVar(&options.Threads, "threads", 3, "Number of threads to start")
	flag.StringVar(&options.OutDir, "out-dir", "", "Directory to clone repositories to")
	flag.BoolVar(&options.SaveRepo, "save-repo", false, "Save the cloned repository")
	flag.BoolVar(&options.NoEmoji, "no-emoji", false, "Disable this if you are a boring person (or use a weird terminal)")
	flag.StringVar(&options.ConfigPath, "config-path", "", "Path to semgrep.conf file")

	// Display options
	flag.BoolVar(&options.Description, "desc", false, "Display repo description")
	flag.BoolVar(&options.FullDescription, "full-desc", false, "Display the full repo description")
	flag.BoolVar(&options.Stars, "stars", false, "Display repos stars in output")
	flag.BoolVar(&options.Language, "lang", false, "Display GitHub repo language")
	flag.BoolVar(&options.Topics, "topics", false, "Display GitHub repo topics")
	flag.BoolVar(&options.Files, "files", false, "Display number of files in repo")
	flag.BoolVar(&options.RepoLink, "repo-link", false, "Display the link associated with the repository")
	flag.BoolVar(&options.Forks, "forks", false, "Display the number of forks of repository")
	flag.BoolVar(&options.LastRelease, "last-release", false, "Display the date of the latest release")
	flag.BoolVar(&options.LastCommit, "last-commit", false, "Display the date of the last commit to the repository")
	flag.BoolVar(&options.FirstCommit, "first-commit", false, "Display the date of the first commit to the repository")
	flag.BoolVar(&options.Commits, "commits", false, "Display the number of commits to the repository")
	flag.BoolVar(&options.SecurityIssues, "security-issues", false, "Display the number of security issues in the repository")
	flag.BoolVar(&options.Watchers, "watchers", false, "Display the number of watchers in a repository")
	flag.BoolVar(&options.Contributors, "contributors", false, "Display the number of contributors in a repository")
	flag.BoolVar(&options.Branch, "branch", false, "Display the default branch of a repository")
	flag.BoolVar(&options.Issues, "issues", false, "Display the number of issues in a repository")
	flag.BoolVar(&options.PullRequests, "pull-requests", false, "Display the number of pull requests in a repository")

	// Filters
	flag.StringVar(&options.FilterStars, "filter-stars", "", "Filter repos stars in output (500-700, -300, 500-, 3000)")
	flag.StringVar(&options.FilterFiles, "filter-files", "", "Filter number of files in repo (500-700, -300, 500-, 3000)")
	flag.StringVar(&options.FilterForks, "filter-forks", "", "Filter the number of forks of repository (500-700, -300, 500-, 3000)")
	flag.StringVar(&options.FilterLastRelease, "filter-last-release", "", "Filter the date of the latest release (yyyy/mm/dd-yyyy/mm/dd, -yyyy/mm/dd, yyyy/mm/dd-, yyyy/mm/dd)")
	flag.StringVar(&options.FilterLastCommit, "filter-last-commit", "", "Filter the date of the last commit to the repository (yyyy/mm/dd-yyyy/mm/dd, -yyyy/mm/dd, yyyy/mm/dd-, yyyy/mm/dd)")
	flag.StringVar(&options.FilterFirstCommit, "filter-first-commit", "", "Filter the date of the first commit to the repository (yyyy/mm/dd-yyyy/mm/dd, -yyyy/mm/dd, yyyy/mm/dd-, yyyy/mm/dd)")
	flag.StringVar(&options.FilterCommits, "filter-commits", "", "Filter the number of commits to the repository (500-700, -300, 500-, 3000)")
	flag.StringVar(&options.FilterSecurityIssues, "filter-security-issues", "", "Filter the number of security issues in the repository (500-700, -300, 500-, 3000)")
	flag.StringVar(&options.FilterWatchers, "filter-watchers", "", "Filter the number of watchers in a repository (500-700, -300, 500-, 3000)")
	flag.StringVar(&options.FilterContributors, "filter-contributors", "", "Filter the number of contributors in a repository (500-700, -300, 500-, 3000)")
	flag.StringVar(&options.FilterIssues, "filter-issues", "", "Filter the number of issues in a repository (500-700, -300, 500-, 3000)")
	flag.StringVar(&options.FilterPullRequests, "filter-pull-requests", "", "Filter the number of pull requests in a repository (500-700, -300, 500-, 3000)")

	// Scanning options
	flag.BoolVar(&options.NoSemgrep, "no-semgrep", false, "Do not perform a semgrep scan on the repos")
	flag.StringVar(&options.SemgrepPath, "semgrep-path", "", "Custom path to the semgrep binary")

	// Reporting options
	flag.BoolVar(&options.Github1s, "github1s", false, "Generate links for the web-based vscode browser at github1s.com rather than github.com")
	flag.BoolVar(&options.RawLinks, "raw-links", false, "Print raw links for semgrep report rather than hyperlink with name, good if you want to save output")

	flag.Parse()

	logger.Configure(options.Debug)

	options.Fireprox = strings.Replace(strings.Replace(options.Fireprox, "http://", "", 1), "https://", "", 1)

	if options.OutDir == "" {
		dirname, err := os.UserHomeDir()
		if err != nil {
			os.Exit(1)
		}

		options.OutDir = filepath.Clean(filepath.Join(dirname, "sastsweep", "scans"))

		if _, err := os.Stat(options.OutDir); os.IsNotExist(err) {
			err = os.MkdirAll(options.OutDir, 0700)
			if err != nil {
				logger.Error("Could not create the sastsweep directory " + options.OutDir + ": " + err.Error())
				os.Exit(1)
			}
		}
	}

	return options
}
