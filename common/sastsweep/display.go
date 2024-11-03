package sastsweep

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chebuya/sastsweep/common"
	"github.com/chebuya/sastsweep/common/logger"

	"github.com/fatih/color"
	"github.com/savioxavier/termlink"
)

const dateStr = "2006-01-03"

var cyan = color.New(color.FgCyan, color.Bold)
var purple = color.New(color.FgMagenta, color.Bold)
var White = color.New(color.FgWhite, color.Bold)
var whiteLite = color.New(color.FgWhite)
var green = color.New(color.FgGreen, color.Bold)

var doPurple = true

func colored(s string) {
	if doPurple {
		purple.Printf(s)
		doPurple = false
	} else {
		cyan.Printf(s)
		doPurple = true
	}
}

func convertToK(stars int) (string, error) {
	if stars >= 1000 {
		return strings.Replace(fmt.Sprintf("%.1fk", float64(stars/1000)), ".0k", "k", 1), nil
	}

	return strconv.Itoa(stars), nil
}

func DisplayRepoInfo(options *common.Options, repoInfo RepoInfo) {
	doPurple = true

	if options.Fireprox != "" {
		White.Printf("%s ", strings.Replace(repoInfo.Target, options.Fireprox, "github.com/", 1))
	} else {
		White.Printf("%s ", repoInfo.Target)
	}

	green.Printf("[200] ")

	if options.Description && repoInfo.Description != "" {
		whiteLite.Printf("[%s] ", repoInfo.Description)
	}

	if options.FullDescription && repoInfo.FullDescription != "" {
		whiteLite.Printf("[%s] ", repoInfo.FullDescription)
	}

	if options.Stars {
		stars, err := convertToK(repoInfo.Stars)
		if err != nil {
			logger.Error("Cannot convert " + strconv.Itoa(repoInfo.Stars) + " into kform")
		} else {
			if options.NoEmoji {
				colored(fmt.Sprintf("[%s stars] ", stars))
			} else {
				colored(fmt.Sprintf("[%s⭐] ", stars))
			}
		}
	}

	if options.Language && repoInfo.Language != "" {
		colored(fmt.Sprintf("[%s] ", strings.Trim(strings.Replace(repoInfo.Language, "%", "%%", -1), ",")))
	}

	if options.Files {
		if options.NoEmoji {
			colored(fmt.Sprintf("[%d files] ", repoInfo.Files))
		} else {
			colored(fmt.Sprintf("[%d📄] ", repoInfo.Files))
		}
	}

	if options.Topics && repoInfo.Topics != "" {
		colored(fmt.Sprintf("[%s] ", strings.Trim(repoInfo.Topics, ",")))
	}

	if options.RepoLink && repoInfo.RepoLink != "" {
		colored(fmt.Sprintf("[%s] ", repoInfo.RepoLink))
	}

	if options.Forks {
		forks, err := convertToK(repoInfo.Forks)
		if err != nil {
			logger.Error("Cannot convert " + strconv.Itoa(repoInfo.Forks) + " into kform")
		} else {
			colored(fmt.Sprintf("[%s forks] ", forks))
		}
	}

	if options.LastRelease && !repoInfo.LastRelease.IsZero() {
		colored(fmt.Sprintf("[last release=%s] ", repoInfo.LastRelease.Format(dateStr)))
	}

	if options.LastCommit {
		colored(fmt.Sprintf("[last commit=%s] ", repoInfo.LastCommit.Format(dateStr)))
	}

	if options.FirstCommit && !repoInfo.FirstCommit.IsZero() {
		colored(fmt.Sprintf("[first commit=%s] ", repoInfo.FirstCommit.Format(dateStr)))
	}

	if options.Commits {
		colored(fmt.Sprintf("[%d commits] ", repoInfo.Commits))
	}

	if options.SecurityIssues {
		colored(fmt.Sprintf("[%d security issues] ", repoInfo.SecurityIssues))
	}

	if options.Watchers {
		colored(fmt.Sprintf("[%d watchers] ", repoInfo.Watchers))
	}

	if options.Contributors {
		colored(fmt.Sprintf("[%d contributors] ", repoInfo.Contributors))
	}

	if options.Branch && repoInfo.Branch != "" {
		colored(fmt.Sprintf("[default branch=%s] ", repoInfo.Branch))
	}

	if options.Issues {
		colored(fmt.Sprintf("[%d issues] ", repoInfo.Issues))
	}

	if options.PullRequests {
		colored(fmt.Sprintf("[%d pull requests] ", repoInfo.PullRequests))
	}

	if !options.NoSemgrep {
		if options.NoEmoji {
			colored(fmt.Sprintf("[%d sg hits]", repoInfo.SemgrepHits))
		} else {
			colored(fmt.Sprintf("[%d🎯]", repoInfo.SemgrepHits))
		}

		fullPath, err := filepath.Abs(repoInfo.ReportPath)
		if err != nil {
			logger.Error("Unable to get the absolute path of " + repoInfo.ReportPath + ": " + err.Error())
		} else if repoInfo.SemgrepHits != 0 {
			if options.RawLinks {
				White.Printf(" [file://" + fullPath + "]")
			} else {
				White.Printf(" [" + termlink.Link("Semgrep report", "file://"+fullPath) + "]")
			}

		}
	}

	fmt.Printf("\n")
}
